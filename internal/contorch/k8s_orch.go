package contorch

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

const defaultNodesMetricsCacheTimeS = 60

type K8sOrchestrator struct {
	config           *rest.Config
	clientset        *kubernetes.Clientset
	metricsClientset *metricsv.Clientset

	nodesStatus    []*model.Node
	nodesCacheTime int
	nodesTime      time.Time
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

	nodesMetricsCacheTimeS, err := strconv.Atoi(os.Getenv("NODE_METRICS_CACHE_TIME_S"))
	if err != nil {
		nodesMetricsCacheTimeS = defaultNodesMetricsCacheTimeS
	}

	return &K8sOrchestrator{
		config:           config,
		clientset:        clientset,
		metricsClientset: metricsClientset,
		nodesCacheTime:   nodesMetricsCacheTimeS,
	}, nil
}

func (orch *K8sOrchestrator) GetNodesStatus() ([]*model.Node, error) {
	if orch.nodesStatus != nil && int(time.Since(orch.nodesTime).Seconds()) < orch.nodesCacheTime {
		log.Println("Using node status cache")
		return orch.nodesStatus, nil
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

	nodes := []*model.Node{}
	for _, nodeCore := range nodesCoreList.Items {
		nodeMetric, exists := nodeMetricsMap[nodeCore.Name]
		if !exists {
			continue
		}

		cpuUsage := nodeMetric.Usage[corev1.ResourceCPU]
		cpuPercentage := float64(cpuUsage.MilliValue()) / float64(nodeCore.Status.Capacity.Cpu().MilliValue())

		memoryUsage := nodeMetric.Usage[corev1.ResourceMemory]
		memoryPercentage := float64(memoryUsage.Value()) / float64(nodeCore.Status.Capacity.Memory().Value())

		hostIP := getHostIp(nodeCore)

		node := &model.Node{
			Id:         nodeCore.Name,
			InternalIp: hostIP,
			Resources: &model.NodeResources{
				CpuUsage: cpuPercentage,
				RamUsage: memoryPercentage,
			},
		}

		nodes = append(nodes, node)
	}

	log.Println("Returning host nodes status ::")
	for _, node := range nodes {
		log.Printf("%+v\n", node)

	}

	orch.nodesStatus = nodes
	orch.nodesTime = time.Now()

	return nodes, nil
}

func (orch *K8sOrchestrator) CreateConfigMapFromFiles(configMapName string, filesData map[string]string) error {
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

func (orch *K8sOrchestrator) CreateDeployment(deployment *appsv1.Deployment) error {
	deploymentsClient := orch.clientset.AppsV1().Deployments(corev1.NamespaceDefault)

	_, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) DeleteDeployment(deploymentName string) error {
	deploymentsClient := orch.clientset.AppsV1().Deployments(corev1.NamespaceDefault)

	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) CreateService(service *corev1.Service) error {
	serviceClient := orch.clientset.CoreV1().Services(corev1.NamespaceDefault)

	_, err := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (orch *K8sOrchestrator) DeleteService(serviceName string) error {
	serviceClient := orch.clientset.CoreV1().Services(corev1.NamespaceDefault)

	deletePolicy := metav1.DeletePropagationForeground
	if err := serviceClient.Delete(context.TODO(), serviceName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return err
	}

	return nil
}

func getHostIp(node corev1.Node) string {
	for _, val := range node.Status.Addresses {
		if val.Type == corev1.NodeInternalIP {
			return val.Address
		}
	}

	return ""
}
