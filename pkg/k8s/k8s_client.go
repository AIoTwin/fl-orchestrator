package k8s

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/pkg/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

const defaultNodesMetricsCacheTimeS = 60

type K8sClient struct {
	config           *rest.Config
	clientset        *kubernetes.Clientset
	metricsClientset *metricsv.Clientset

	nodesStatus    []*model.Node
	nodesCacheTime int
	nodesTime      time.Time
}

func NewK8sClient(configFilePath string) (*K8sClient, error) {
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
	log.Println("NODE_METRICS_CACHE_TIME_S:", nodesMetricsCacheTimeS)

	return &K8sClient{
		config:           config,
		clientset:        clientset,
		metricsClientset: metricsClientset,
		nodesCacheTime:   nodesMetricsCacheTimeS,
	}, nil
}

func (c *K8sClient) GetNodesStatus() ([]*model.Node, error) {
	if c.nodesStatus != nil && int(time.Since(c.nodesTime).Seconds()) < c.nodesCacheTime {
		log.Println("Using node status cache")
		return c.nodesStatus, nil
	}

	nodesCoreList, err := c.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println("Failed to retrieve nodes on node status")
		return nil, err
	}

	nodeMetricsList, err := c.metricsClientset.MetricsV1beta1().NodeMetricses().List(context.Background(), metav1.ListOptions{})
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

	c.nodesStatus = nodes
	c.nodesTime = time.Now()

	return nodes, nil
}

func getHostIp(node corev1.Node) string {
	for _, val := range node.Status.Addresses {
		if val.Type == corev1.NodeInternalIP {
			return val.Address
		}
	}

	return ""
}
