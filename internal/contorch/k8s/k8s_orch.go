package k8sorch

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
const communicationCostPrefix = "comm/"

type K8sOrchestrator struct {
	config           *rest.Config
	clientset        *kubernetes.Clientset
	metricsClientset *metricsv.Clientset
	eventBus         *events.EventBus
	cronScheduler    *cron.Cron
	availableNodes   map[string]*model.Node
	simulation       bool
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
	}, nil
}

func (orch *K8sOrchestrator) GetAvailableNodes(initialRequest bool) (map[string]*model.Node, error) {
	if orch.simulation {
		nodes := common.GetAvailableNodesFromFile()
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

func (orch *K8sOrchestrator) notifyNodeStateChanges() {
	availableNodesNew, err := orch.GetAvailableNodes(false)
	if err != nil {
		return
	}

	// check for removed nodes
	for _, node := range orch.availableNodes {
		_, found := availableNodesNew[node.Id]
		if !found {
			// Create the user registered event
			event := events.Event{
				Type:      common.NODE_STATE_CHANGE_EVENT_TYPE,
				Timestamp: time.Now(),
				Data: events.NodeStateChangeEvent{
					State: common.NODE_REMOVED,
					Node:  node,
				},
			}

			orch.eventBus.Publish(event)
		}
	}

	// check for added nodes
	for _, node := range availableNodesNew {
		_, found := orch.availableNodes[node.Id]
		if !found {
			// Create the user registered event
			event := events.Event{
				Type:      common.NODE_STATE_CHANGE_EVENT_TYPE,
				Timestamp: time.Now(),
				Data: events.NodeStateChangeEvent{
					State: common.NODE_ADDED,
					Node:  node,
				},
			}

			orch.eventBus.Publish(event)
		}
	}

	orch.availableNodes = availableNodesNew
}

func (orch *K8sOrchestrator) CreateGlobalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error {
	err := orch.createConfigMapFromFiles(common.GetAggregatorConfigMapName(aggregator.Id), configFiles)
	if err != nil {
		return err
	}

	deployment := BuildGlobalAggregatorDeployment(aggregator)
	if !orch.simulation {
		deployment.Spec.Template.Spec.NodeName = aggregator.Id
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

func (orch *K8sOrchestrator) RemoveGlobalAggregator(aggregator *model.FlAggregator) error {
	err := orch.deleteService(common.GetAggregatorServiceName(aggregator.Id))
	if err != nil {
		return err
	}

	err = orch.deleteDeployment(common.GLOBAL_AGGRETATOR_DEPLOYMENT_NAME)
	if err != nil {
		return err
	}

	err = orch.deleteConfigMap(common.GetAggregatorConfigMapName(aggregator.Id))
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
		FlType:             getFlType(nodeCore.Labels),
		CommunicationCosts: getCommunicationCosts(nodeCore.Labels),
	}

	return nodeModel
}

func getFlType(labels map[string]string) string {
	flType := labels[flTypeLabel]
	return flType
}

func getCommunicationCosts(labels map[string]string) map[string]float32 {
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

	return communicationCosts
}

func getHostIp(node corev1.Node) string {
	for _, val := range node.Status.Addresses {
		if val.Type == corev1.NodeInternalIP {
			return val.Address
		}
	}

	return ""
}
