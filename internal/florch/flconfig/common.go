package flconfig

import (
	"fmt"
	"math"
	"strconv"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

func validPartition(clusters [][]*model.Node, clusterSizes []int) bool {
	for i, cluster := range clusters {
		if len(cluster) != clusterSizes[i] {
			return false
		}
	}
	return true
}

func printClusters(clusters [][]*model.Node) {
	for _, cluster := range clusters {
		fmt.Print("[")
		for i, node := range cluster {
			if i != 0 {
				fmt.Print(" ")
			}
			fmt.Printf("%s", node.Id)
		}
		fmt.Print("] ")
	}
	fmt.Println()
}

func getTotalKld(clusters [][]*model.Node, averageDistribution []float64) float64 {
	klds := make([]float64, len(clusters))
	for i, cluster := range clusters {
		clusterDataDistribution := getClusterDataDistribution(cluster)
		klds[i] = klDivergence(clusterDataDistribution, averageDistribution)
	}

	return calculateAverage(klds)
}

func getClusterDataDistribution(nodes []*model.Node) []float64 {
	totalSamples := 0
	samplesPerClass := make([]int64, 10)
	for _, node := range nodes {
		if node.FlType == common.FL_TYPE_CLIENT {
			dataDistribution := node.DataDistribution
			for class, samples := range dataDistribution {
				i, _ := strconv.Atoi(class)
				samplesPerClass[i] += samples
				totalSamples += int(samples)
			}
		}
	}

	clusterDistribution := make([]float64, 10)
	for i, samples := range samplesPerClass {
		percentage := float64(samples) / float64(totalSamples)
		if percentage == 0.0 {
			percentage = 0.0001
		}
		clusterDistribution[i] = percentage
	}

	return clusterDistribution
}

func klDivergence(p, q []float64) float64 {
	if len(p) != len(q) {
		panic("Distributions must have the same number of parameters")
	}

	klDiv := 0.0
	for i := 0; i < len(p); i++ {
		if q[i] == 0 {
			continue
		}
		klDiv += p[i] * math.Log(p[i]/q[i])
	}
	return klDiv
}

func calculateAverage(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	var sum float64
	for _, number := range numbers {
		sum += number
	}

	return sum / float64(len(numbers))
}

func getHierarchicalAggregationCosts(globalAggregator *model.Node, localAggregators []*model.Node, clusters [][]*model.Node,
	modelSize float32) (float32, float32, error) {
	globalAggregationCost, err := calculateAggregationCost(localAggregators, globalAggregator.Id, modelSize)
	if err != nil {
		return 0.0, 0.0, nil
	}

	localAggregationCost := float32(0)
	for i, cluster := range clusters {
		clusterAggregationCost, err := calculateAggregationCost(cluster, localAggregators[i].Id, modelSize)
		if err != nil {
			return 0.0, 0.0, nil
		}

		localAggregationCost += clusterAggregationCost
	}

	return globalAggregationCost, localAggregationCost, nil
}

func calculateAggregationCost(clients []*model.Node, aggregatorNodeId string, modelSize float32) (float32, error) {
	aggregationCost := float32(0.0)
	for _, client := range clients {
		communicationCosts := client.CommunicationCosts
		cost, exists := communicationCosts[aggregatorNodeId]
		if !exists {
			return 0.0, fmt.Errorf("no comm cost value from client %s to aggregator %s", client.Id, aggregatorNodeId)
		}
		aggregationCost += cost * modelSize
	}

	return aggregationCost, nil
}

func (config *CentrHierFlConfiguration) partitionClients(clients []*model.Node, index int, clusters [][]*model.Node, clusterSizes []int) {
	if index == len(clients) {
		if validPartition(clusters, clusterSizes) {
			kld := getTotalKld(clusters, config.averageDistribution)
			if kld < config.bestKld {
				config.bestKld = kld
				config.bestClusters = make([][]*model.Node, len(clusters))
				copy(config.bestClusters, clusters)
			}
		}
		return
	}

	for i := 0; i < len(clusters); i++ {
		if len(clusters[i]) < clusterSizes[i] {
			newClusters := make([][]*model.Node, len(clusters))
			copy(newClusters, clusters)
			newClusters[i] = append(newClusters[i], clients[index])
			config.partitionClients(clients, index+1, newClusters, clusterSizes)
		}
	}
}
