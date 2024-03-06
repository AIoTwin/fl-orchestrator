package k8sorch

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
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
}

func NewK8sOrchestrator(configFilePath string) (*K8sOrchestrator, error) {
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
	}, nil
}

func (orch *K8sOrchestrator) GetAvailableNodes() ([]*model.Node, error) {
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

	nodes := []*model.Node{}
	for _, nodeCore := range nodesCoreList.Items {
		nodeMetric, exists := nodeMetricsMap[nodeCore.Name]
		if !exists {
			continue
		}

		if !isNodeReady(nodeCore) {
			continue
		}

		nodeModel := nodeCoreToNodeModel(nodeCore, nodeMetric)

		nodes = append(nodes, nodeModel)
	}

	log.Println("Returning nodes ::")
	for _, node := range nodes {
		log.Printf("%+v\n", node)
	}

	return nodes, nil
}

func (orch *K8sOrchestrator) CreateGlobalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error {
	err := orch.createConfigMapFromFiles(common.GetAggregatorConfigMapName(aggregator.Id), configFiles)
	if err != nil {
		return err
	}

	deployment := BuildGlobalAggregatorDeployment(aggregator)
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

func (orch *K8sOrchestrator) CreateFlClient(client *model.FlClient, configFiles map[string]string) error {
	err := orch.createConfigMapFromFiles(common.GetClientConfigMapName(client.Id), configFiles)
	if err != nil {
		return err
	}

	deployment := BuildClientDeployment(client)
	err = orch.createDeployment(deployment)
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) createConfigMapFromFiles(configMapName string, filesData map[string]string) error {
	// Create a ConfigMap object
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: corev1.NamespaceDefault,
		},
		Data: filesData,
	}

	// Create or update the ConfigMap
	_, err := orch.clientset.CoreV1().ConfigMaps(corev1.NamespaceDefault).Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating ConfigMap: %v\n", err)
		return err
	}
	fmt.Println("ConfigMap created successfully.")

	return nil
}

func (orch *K8sOrchestrator) createDeployment(deployment *appsv1.Deployment) error {
	deploymentsClient := orch.clientset.AppsV1().Deployments(corev1.NamespaceDefault)

	_, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})

	return err
}

func (orch *K8sOrchestrator) deleteDeployment(deploymentName string) error {
	deploymentsClient := orch.clientset.AppsV1().Deployments(corev1.NamespaceDefault)

	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) createService(service *corev1.Service) error {
	serviceClient := orch.clientset.CoreV1().Services(corev1.NamespaceDefault)

	_, err := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) deleteService(serviceName string) error {
	serviceClient := orch.clientset.CoreV1().Services(corev1.NamespaceDefault)

	deletePolicy := metav1.DeletePropagationForeground
	if err := serviceClient.Delete(context.TODO(), serviceName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
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
