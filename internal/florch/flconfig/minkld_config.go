package flconfig

import (
	"fmt"
	"math"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type MinimizeKldConfiguration struct {
	epochs              int32
	localRounds         int32
	bestClusters        [][]*model.FlClient
	averageDistribution []float64
	bestKld             float64
}

func NewMinimizeKldConfiguration(epochs int32, localRounds int32) *MinimizeKldConfiguration {
	return &MinimizeKldConfiguration{
		epochs:      epochs,
		localRounds: localRounds,
	}
}

func (config *MinimizeKldConfiguration) GetOptimalConfiguration(flEntitiesInitial *model.FlEntities) *FlConfiguration {
	flEntities := copyFlEntitites(flEntitiesInitial)

	if len(flEntities.LocalAggregators) <= 1 {
		// centralized
		flEntities.GlobalAggregator.NumClients = int32(len(flEntities.Clients))
		for _, client := range flEntities.Clients {
			client.Epochs = config.epochs * config.localRounds
			client.ParentAddress = flEntities.GlobalAggregator.ExternalAddress
			client.ParentNodeId = flEntities.GlobalAggregator.Id
		}

		return &FlConfiguration{
			FlEntities: flEntities,
			Epochs:     config.epochs * config.localRounds,
		}
	}

	// note: this is simple example of clustering with equal distribution of clients per aggregator
	config.bestClusters = make([][]*model.FlClient, 0)
	config.averageDistribution = make([]float64, 0)
	config.bestKld = math.MaxFloat64

	// get cluster sizes
	numClients := len(flEntities.Clients)
	numLocalAggregators := len(flEntities.LocalAggregators)
	div := numClients / numLocalAggregators
	mod := numClients % numLocalAggregators
	clusters := make([][]*model.FlClient, numLocalAggregators)
	clusterSizes := make([]int, numLocalAggregators)
	for i := 0; i < numLocalAggregators; i++ {
		if i < mod {
			clusterSizes[i] = div + 1
		} else {
			clusterSizes[i] = div
		}
	}

	// make optimal clusters
	config.averageDistribution = getClusterDataDistribution(flEntities.Clients)
	config.bestKld = math.MaxFloat64
	config.partitionClients(flEntities.Clients, 0, clusters, clusterSizes)
	fmt.Print("Optimal clusters: ")
	printClusters(config.bestClusters)
	fmt.Println("Best KLD: ", config.bestKld)

	// prepare clients and aggregators
	flEntities.GlobalAggregator.NumClients = int32(len(flEntities.LocalAggregators))
	clients := []*model.FlClient{}
	for n, cluster := range config.bestClusters {
		localAggregator := flEntities.LocalAggregators[n]
		localAggregator.LocalRounds = config.localRounds
		localAggregator.ParentAddress = flEntities.GlobalAggregator.ExternalAddress
		flClientsCluster := common.PrepareFlClients(cluster, localAggregator, config.epochs)
		clients = append(clients, flClientsCluster...)
	}
	flEntities.Clients = clients

	return &FlConfiguration{
		FlEntities:  flEntities,
		Epochs:      config.epochs,
		LocalRounds: config.localRounds,
	}
}

func (config *MinimizeKldConfiguration) partitionClients(clients []*model.FlClient, index int, clusters [][]*model.FlClient, clusterSizes []int) {
	if index == len(clients) {
		if validPartition(clusters, clusterSizes) {
			kld := getTotalKld(clusters, config.averageDistribution)
			if kld < config.bestKld {
				config.bestKld = kld
				config.bestClusters = deepCopyClusters(clusters)
			}
		}
		return
	}

	for i := 0; i < len(clusters); i++ {
		if len(clusters[i]) < clusterSizes[i] {
			// Create a deep copy of the clusters for each recursive call
			newClusters := deepCopyClusters(clusters)
			newClusters[i] = append(newClusters[i], clients[index])
			config.partitionClients(clients, index+1, newClusters, clusterSizes)
		}
	}
}

// deepCopyClusters creates a deep copy of the clusters slice
func deepCopyClusters(clusters [][]*model.FlClient) [][]*model.FlClient {
	newClusters := make([][]*model.FlClient, len(clusters))
	for i := range clusters {
		newClusters[i] = make([]*model.FlClient, len(clusters[i]))
		copy(newClusters[i], clusters[i])
	}
	return newClusters
}
