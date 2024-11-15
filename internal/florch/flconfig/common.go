package flconfig

import (
	"fmt"
	"math"
	"strconv"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

func copyFlEntitites(srcFlEntities *model.FlEntities) *model.FlEntities {
	localAggregators := []*model.FlAggregator{}
	clients := []*model.FlClient{}

	for _, la := range srcFlEntities.LocalAggregators {
		localAggregators = append(localAggregators, copyFlAggregator(la))
	}

	for _, cl := range srcFlEntities.Clients {
		clients = append(clients, copyFlClient(cl))
	}

	return &model.FlEntities{
		GlobalAggregator: copyFlAggregator(srcFlEntities.GlobalAggregator),
		LocalAggregators: localAggregators,
		Clients:          clients,
	}
}

func copyFlAggregator(srcGlobalAggregator *model.FlAggregator) *model.FlAggregator {
	return &model.FlAggregator{
		Id:                 srcGlobalAggregator.Id,
		InternalAddress:    srcGlobalAggregator.InternalAddress,
		ExternalAddress:    srcGlobalAggregator.ExternalAddress,
		ParentAddress:      srcGlobalAggregator.ParentAddress,
		Port:               srcGlobalAggregator.Port,
		NumClients:         srcGlobalAggregator.NumClients,
		Rounds:             srcGlobalAggregator.Rounds,
		LocalRounds:        srcGlobalAggregator.LocalRounds,
		CommunicationCosts: srcGlobalAggregator.CommunicationCosts,
	}
}

func copyFlClient(srcFlClient *model.FlClient) *model.FlClient {
	return &model.FlClient{
		Id:                 srcFlClient.Id,
		ParentAddress:      srcFlClient.ParentAddress,
		ParentNodeId:       srcFlClient.ParentNodeId,
		Epochs:             srcFlClient.Epochs,
		CommunicationCosts: srcFlClient.CommunicationCosts,
		DataDistribution:   srcFlClient.DataDistribution,
		ClientUtility:      srcFlClient.ClientUtility,
	}
}

func validPartition(clusters [][]*model.FlClient, clusterSizes []int) bool {
	for i, cluster := range clusters {
		if len(cluster) != clusterSizes[i] {
			return false
		}
	}
	return true
}

func printClusters(clusters [][]*model.FlClient) {
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

func getTotalKld(clusters [][]*model.FlClient, averageDistribution []float64) float64 {
	klds := make([]float64, len(clusters))
	for i, cluster := range clusters {
		clusterDataDistribution := getClusterDataDistribution(cluster)
		klds[i] = klDivergence(clusterDataDistribution, averageDistribution)
	}

	return common.CalculateAverageFloat64(klds)
}

func getClusterDataDistribution(clients []*model.FlClient) []float64 {
	totalSamples := 0
	samplesPerClass := make([]int64, 10)
	for _, client := range clients {
		dataDistribution := client.DataDistribution
		for class, samples := range dataDistribution {
			i, _ := strconv.Atoi(class)
			samplesPerClass[i] += samples
			totalSamples += int(samples)
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

func getHierarchicalAggregationCosts(globalAggregator *model.FlAggregator, localAggregators []*model.FlAggregator, clusters [][]*model.FlClient,
	modelSize float32) (float32, float32, error) {
	globalAggregationCost, err := calculateGlobalAggregationCost(localAggregators, globalAggregator.Id, modelSize)
	if err != nil {
		return 0.0, 0.0, nil
	}

	localAggregationCost := float32(0)
	for i, cluster := range clusters {
		clusterAggregationCost, err := calculateLocalAggregationCost(cluster, localAggregators[i].Id, modelSize)
		if err != nil {
			return 0.0, 0.0, nil
		}

		localAggregationCost += clusterAggregationCost
	}

	return globalAggregationCost, localAggregationCost, nil
}

func calculateLocalAggregationCost(clients []*model.FlClient, localAggregatorNodeId string, modelSize float32) (float32, error) {
	aggregationCost := float32(0.0)
	for _, client := range clients {
		communicationCosts := client.CommunicationCosts
		cost, exists := communicationCosts[localAggregatorNodeId]
		if !exists {
			return 0.0, fmt.Errorf("no comm cost value from client %s to aggregator %s", client.Id, localAggregatorNodeId)
		}
		aggregationCost += cost * modelSize
	}

	return aggregationCost, nil
}

func calculateGlobalAggregationCost(localAggregators []*model.FlAggregator, globalAggregatorNodeId string, modelSize float32) (float32, error) {
	aggregationCost := float32(0.0)
	for _, localAggregator := range localAggregators {
		communicationCosts := localAggregator.CommunicationCosts
		cost, exists := communicationCosts[globalAggregatorNodeId]
		if !exists {
			return 0.0, fmt.Errorf("no comm cost value from LA %s to GA %s", localAggregator.Id, globalAggregatorNodeId)
		}
		aggregationCost += cost * modelSize
	}

	return aggregationCost, nil
}

func (config *CentrHierFlConfiguration) partitionClients(clients []*model.FlClient, index int, clusters [][]*model.FlClient, clusterSizes []int) {
	if index == len(clients) {
		if validPartition(clusters, clusterSizes) {
			kld := getTotalKld(clusters, config.averageDistribution)
			if kld < config.bestKld {
				config.bestKld = kld
				config.bestClusters = make([][]*model.FlClient, len(clusters))
				copy(config.bestClusters, clusters)
			}
		}
		return
	}

	for i := 0; i < len(clusters); i++ {
		if len(clusters[i]) < clusterSizes[i] {
			newClusters := make([][]*model.FlClient, len(clusters))
			copy(newClusters, clusters)
			newClusters[i] = append(newClusters[i], clients[index])
			config.partitionClients(clients, index+1, newClusters, clusterSizes)
		}
	}
}
