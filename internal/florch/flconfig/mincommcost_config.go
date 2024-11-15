package flconfig

import (
	"fmt"
	"math"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type MinimizeCommCostConfiguration struct {
	epochs       int32
	localRounds  int32
	modelSize    float32
	bestClusters [][]*model.FlClient
	bestCommCost float32
	flEntities   *model.FlEntities
}

func NewMinimizeCommCostConfiguration(epochs int32, localRounds int32, modelSize float32) *MinimizeCommCostConfiguration {
	return &MinimizeCommCostConfiguration{
		epochs:      epochs,
		localRounds: localRounds,
		modelSize:   modelSize,
	}
}

func (config *MinimizeCommCostConfiguration) GetOptimalConfiguration(flEntitiesInitial *model.FlEntities) *FlConfiguration {
	config.flEntities = copyFlEntitites(flEntitiesInitial)

	if len(config.flEntities.LocalAggregators) <= 1 {
		// centralized
		config.flEntities.GlobalAggregator.NumClients = int32(len(config.flEntities.Clients))
		config.flEntities.Clients = common.PrepareFlClients(config.flEntities.Clients, config.flEntities.GlobalAggregator, config.epochs*config.localRounds)

		return &FlConfiguration{
			FlEntities: config.flEntities,
			Epochs:     config.epochs * config.localRounds,
		}
	}

	// note: this is simple example of clustering with equal distribution of clients per aggregator
	config.bestClusters = make([][]*model.FlClient, 0)

	// get cluster sizes
	numClients := len(config.flEntities.Clients)
	numLocalAggregators := len(config.flEntities.LocalAggregators)
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
	config.bestCommCost = math.MaxFloat32
	config.partitionClients(config.flEntities.Clients, 0, clusters, clusterSizes)
	fmt.Print("Optimal clusters: ")
	printClusters(config.bestClusters)
	fmt.Println("Best comm cost: ", config.bestCommCost)

	// prepare clients and aggregators
	config.flEntities.GlobalAggregator.NumClients = int32(len(config.flEntities.LocalAggregators))
	clients := []*model.FlClient{}
	for n, cluster := range config.bestClusters {
		localAggregator := config.flEntities.LocalAggregators[n]
		localAggregator.LocalRounds = config.localRounds
		localAggregator.ParentAddress = config.flEntities.GlobalAggregator.ExternalAddress
		flClientsCluster := common.PrepareFlClients(cluster, localAggregator, config.epochs)
		clients = append(clients, flClientsCluster...)
	}
	config.flEntities.Clients = clients

	return &FlConfiguration{
		FlEntities:  config.flEntities,
		Epochs:      config.epochs,
		LocalRounds: config.localRounds,
	}
}

func (config *MinimizeCommCostConfiguration) partitionClients(clients []*model.FlClient, index int, clusters [][]*model.FlClient, clusterSizes []int) {
	if index == len(clients) {
		if validPartition(clusters, clusterSizes) {
			gaCost, laCost, _ := getHierarchicalAggregationCosts(config.flEntities.GlobalAggregator, config.flEntities.LocalAggregators, clusters,
				config.modelSize)
			commCost := gaCost + laCost
			if commCost < config.bestCommCost {
				config.bestCommCost = commCost
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
