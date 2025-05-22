package k8sorch

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"github.com/robfig/cron/v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

const flTypeLabel = "fl/type"
const numPartitionsLabel = "fl/num-partitions"
const partitionIdLabel = "fl/partition-id"
const communicationCostPrefix = "comm/"
const dataDistributionPrefix = "data/"

type K8sOrchestrator struct {
	config             *rest.Config
	clientset          *kubernetes.Clientset
	metricsClientset   *metricsv.Clientset
	eventBus           *events.EventBus
	cronScheduler      *cron.Cron
	availableNodes     map[string]*model.Node
	simulation         bool
	simulationNodes    []string
	lastSimulationNode int
}

func NewK8sOrchestrator(configFilePath string, eventBus *events.EventBus, simulation bool) (*K8sOrchestrator, error) {
	// connect to Kubernetes cluster
	config, err := clientcmd.BuildConfigFromFlags("", configFilePath)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &K8sOrchestrator{
		config:           config,
		clientset:        clientset,
		metricsClientset: metricsClientset,
		eventBus:         eventBus,
		cronScheduler:    cron.New(cron.WithSeconds()),
		availableNodes:   make(map[string]*model.Node),
		simulation:       simulation,
		simulationNodes: []string{"fl-k3s-ga", "fl-k3s-la-1", "fl-k3s-la-2", "fl-k3s-node-1", "fl-k3s-node-2",
			"fl-k3s-node-3", "fl-k3s-node-4", "fl-k3s-node-5", "fl-k3s-node-6", "fl-k3s-node-7", "fl-k3s-node-8"},
		lastSimulationNode: 0,
	}, nil
}

func (orch *K8sOrchestrator) GetAvailableNodes(initialRequest bool) (map[string]*model.Node, error) {
	if orch.simulation {
		nodes, err := common.GetAvailableNodesFromFile()
		if err != nil {
			return nil, err
		}
		if initialRequest {
			for _, node := range nodes {
				orch.availableNodes[node.Id] = node
			}
		}

		return nodes, nil
	}

	nodesCoreList, err := orch.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println("Failed to retrieve nodes on node status")
		return nil, err
	}

	nodeMetricsList, err := orch.metricsClientset.MetricsV1beta1().NodeMetricses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println("Failed to retrieve node metrices on node status")
		return nil, err
	}

	nodeMetricsMap := make(map[string]v1beta1.NodeMetrics)
	for _, nodeMetric := range nodeMetricsList.Items {
		nodeMetricsMap[nodeMetric.Name] = nodeMetric
	}

	nodes := make(map[string]*model.Node)
	for _, nodeCore := range nodesCoreList.Items {
		nodeMetric, exists := nodeMetricsMap[nodeCore.Name]
		if !exists {
			continue
		}

		if !isNodeReady(nodeCore) {
			continue
		}

		nodeModel := nodeCoreToNodeModel(nodeCore, nodeMetric)

		nodes[nodeModel.Id] = nodeModel

		if initialRequest {
			orch.availableNodes[nodeModel.Id] = nodeModel
		}
	}

	return nodes, nil
}

func (orch *K8sOrchestrator) StartNodeStateChangeNotifier() {
	orch.cronScheduler.AddFunc("@every 1s", orch.notifyNodeStateChanges)

	orch.cronScheduler.Start()
}

func (orch *K8sOrchestrator) StopAllNotifiers() {
	orch.cronScheduler.Stop()
}

func (orch *K8sOrchestrator) notifyNodeStateChanges() {
	availableNodesNew, err := orch.GetAvailableNodes(false)
	if err != nil {
		return
	}

	event := common.GetNodeStateChangeEvent(orch.availableNodes, availableNodesNew)
	if (event != events.Event{}) {
		orch.eventBus.Publish(event)
	}

	orch.availableNodes = availableNodesNew
}

func (orch *K8sOrchestrator) CreateGlobalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error {
	err := orch.createConfigMapFromFiles(common.GetGlobalAggregatorConfigMapName(aggregator.Id), configFiles)
	if err != nil {
		return err
	}

	deployment := BuildGlobalAggregatorDeployment(aggregator)
	if !orch.simulation {
		deployment.Spec.Template.Spec.NodeName = aggregator.Id
	} else {
		deployment.Spec.Template.Spec.NodeName = orch.simulationNodes[orch.lastSimulationNode]
		orch.lastSimulationNode++
	}

	err = orch.createDeployment(deployment)
	if err != nil {
		return err
	}

	service := BuildGlobalAggregatorService(aggregator)
	err = orch.createService(service)
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) GetGlobalAggregatorLogs() (bytes.Buffer, error) {
	// Get the deployment
	deployment, err := orch.clientset.AppsV1().Deployments(corev1.NamespaceDefault).Get(context.TODO(),
		common.GLOBAL_AGGRETATOR_DEPLOYMENT_NAME, metav1.GetOptions{})
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error retrieving deployment: %v", err)
	}

	// Get the selector from the deployment
	labelSelector := metav1.FormatLabelSelector(deployment.Spec.Selector)

	// List pods with the same labels
	podList, err := orch.clientset.CoreV1().Pods(corev1.NamespaceDefault).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error listing pods: %v", err)
	}

	// Get the logs of the pod
	req := orch.clientset.CoreV1().Pods(corev1.NamespaceDefault).GetLogs(podList.Items[0].Name, &corev1.PodLogOptions{})
	logs, err := req.Stream(context.TODO())
	if err != nil {
		return bytes.Buffer{}, err
	}
	defer logs.Close()

	// Read the logs into a buffer
	var buf bytes.Buffer
	_, err = buf.ReadFrom(logs)
	if err != nil {
		return bytes.Buffer{}, err
	}

	return buf, nil
}

func (orch *K8sOrchestrator) RemoveGlobalAggregator(aggregator *model.FlAggregator) error {
	err := orch.deleteService(common.GetGlobalAggregatorServiceName(aggregator.Id))
	if err != nil {
		return err
	}

	err = orch.deleteDeployment(common.GLOBAL_AGGRETATOR_DEPLOYMENT_NAME)
	if err != nil {
		return err
	}

	err = orch.deleteConfigMap(common.GetGlobalAggregatorConfigMapName(aggregator.Id))
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) CreateLocalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error {
	err := orch.createConfigMapFromFiles(common.GetLocalAggregatorConfigMapName(aggregator.Id), configFiles)
	if err != nil {
		return err
	}

	deployment := BuildLocalAggregatorDeployment(aggregator)
	if !orch.simulation {
		deployment.Spec.Template.Spec.NodeName = aggregator.Id
	} else {
		deployment.Spec.Template.Spec.NodeName = orch.simulationNodes[orch.lastSimulationNode]
		orch.lastSimulationNode++
	}

	err = orch.createDeployment(deployment)
	if err != nil {
		return err
	}

	service := BuildLocalAggregatorService(aggregator)
	err = orch.createService(service)
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) RemoveLocalAggregator(aggregator *model.FlAggregator) error {
	err := orch.deleteService(common.GetLocalAggregatorServiceName(aggregator.Id))
	if err != nil {
		return err
	}

	err = orch.deleteDeployment(common.GetLocalAggregatorDeploymentName(aggregator.Id))
	if err != nil {
		return err
	}

	err = orch.deleteConfigMap(common.GetLocalAggregatorConfigMapName(aggregator.Id))
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) CreateFlClient(client *model.FlClient, configFiles map[string]string) error {
	err := orch.createConfigMapFromFiles(common.GetClientConfigMapName(client.Id), configFiles)
	if err != nil {
		return err
	}

	deployment := BuildClientDeployment(client)
	if !orch.simulation {
		deployment.Spec.Template.Spec.NodeName = client.Id
	} else {
		deployment.Spec.Template.Spec.NodeName = orch.simulationNodes[orch.lastSimulationNode]
		orch.lastSimulationNode++
		if orch.lastSimulationNode == len(orch.simulationNodes) {
			orch.lastSimulationNode = 3
		}
	}

	err = orch.createDeployment(deployment)
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) RemoveClient(client *model.FlClient) error {
	err := orch.deleteDeployment(common.GetClientDeploymentName(client.Id))
	if err != nil {
		return err
	}

	err = orch.deleteConfigMap(common.GetClientConfigMapName(client.Id))
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) createConfigMapFromFiles(configMapName string, filesData map[string]string) error {
	configMapsClient := orch.clientset.CoreV1().ConfigMaps(corev1.NamespaceDefault)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: corev1.NamespaceDefault,
		},
		Data: filesData,
	}

	_, err := configMapsClient.Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating ConfigMap: %v\n", err)
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) deleteConfigMap(configMapName string) error {
	configMapsClient := orch.clientset.CoreV1().ConfigMaps(corev1.NamespaceDefault)

	if err := configMapsClient.Delete(context.TODO(), configMapName, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) createDeployment(deployment *appsv1.Deployment) error {
	deploymentsClient := orch.clientset.AppsV1().Deployments(corev1.NamespaceDefault)

	_, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})

	return err
}

func (orch *K8sOrchestrator) deleteDeployment(deploymentName string) error {
	deploymentsClient := orch.clientset.AppsV1().Deployments(corev1.NamespaceDefault)

	if err := deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) createService(service *corev1.Service) error {
	servicesClient := orch.clientset.CoreV1().Services(corev1.NamespaceDefault)

	_, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) deleteService(serviceName string) error {
	servicesClient := orch.clientset.CoreV1().Services(corev1.NamespaceDefault)

	if err := servicesClient.Delete(context.TODO(), serviceName, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

// HELPER METHODS

func isNodeReady(nodeCore corev1.Node) bool {
	for _, condition := range nodeCore.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

func nodeCoreToNodeModel(nodeCore corev1.Node, nodeMetric v1beta1.NodeMetrics) *model.Node {
	cpuUsage := nodeMetric.Usage[corev1.ResourceCPU]
	cpuPercentage := float64(cpuUsage.MilliValue()) / float64(nodeCore.Status.Capacity.Cpu().MilliValue())

	memoryUsage := nodeMetric.Usage[corev1.ResourceMemory]
	memoryPercentage := float64(memoryUsage.Value()) / float64(nodeCore.Status.Capacity.Memory().Value())

	hostIP := getHostIp(nodeCore)

	nodeModel := &model.Node{
		Id:         nodeCore.Name,
		InternalIp: hostIP,
		Resources: model.NodeResources{
			CpuUsage: cpuPercentage,
			RamUsage: memoryPercentage,
		},
	}

	nodeLabelsToNodeModel(nodeCore.Labels, nodeModel)

	return nodeModel
}

func nodeLabelsToNodeModel(labels map[string]string, nodeModel *model.Node) {
	flType := labels[flTypeLabel]
	numPartitions, _ := strconv.Atoi(labels[numPartitionsLabel])
	partitionId, _ := strconv.Atoi(labels[partitionIdLabel])
	communicationCosts := make(map[string]float32)
	for key, value := range labels {
		if strings.HasPrefix(key, communicationCostPrefix) {
			splits := strings.Split(key, communicationCostPrefix)
			if len(splits) == 2 {
				cost, _ := strconv.ParseFloat(value, 32)
				communicationCosts[splits[1]] = float32(cost)
			}
		}
	}

	nodeModel.FlType = flType
	nodeModel.CommunicationCosts = communicationCosts
	nodeModel.NumPartitions = int32(numPartitions)
	nodeModel.PartitionId = int32(partitionId)
}

func getFlType(labels map[string]string) string {
	flType := labels[flTypeLabel]
	return flType
}

func getCommCostsAndDataDistribution(labels map[string]string) (map[string]float32, map[string]int64) {
	communicationCosts := make(map[string]float32)
	dataDistribution := make(map[string]int64)
	for key, value := range labels {
		if strings.HasPrefix(key, communicationCostPrefix) {
			splits := strings.Split(key, communicationCostPrefix)
			if len(splits) == 2 {
				cost, _ := strconv.ParseFloat(value, 32)
				communicationCosts[splits[1]] = float32(cost)
			}
		} else if strings.HasPrefix(key, dataDistributionPrefix) {
			splits := strings.Split(key, dataDistributionPrefix)
			if len(splits) == 2 {
				numberOfSamples, _ := strconv.Atoi(value)
				dataDistribution[splits[1]] = int64(numberOfSamples)
			}
		}
	}

	return communicationCosts, dataDistribution
}

func getHostIp(node corev1.Node) string {
	for _, val := range node.Status.Addresses {
		if val.Type == corev1.NodeInternalIP {
			return val.Address
		}
	}

	return ""
}
