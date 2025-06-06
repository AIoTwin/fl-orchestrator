package flconfig

import (
	"fmt"
	"math"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type MinimizeCommCostGreedyConfiguration struct {
	epochs               int32
	localRounds          int32
	modelSize            float32
	bestClusters         [][]*model.Node
	bestCommCost         float32
	globalAgregatorNode  *model.Node
	localAggregatorNodes []*model.Node
}

func NewMinimizeCommCostGreedyConfiguration(epochs int32, localRounds int32, modelSize float32) *MinimizeCommCostGreedyConfiguration {
	return &MinimizeCommCostGreedyConfiguration{
		epochs:      epochs,
		localRounds: localRounds,
		modelSize:   modelSize,
	}
}

func (config *MinimizeCommCostGreedyConfiguration) GetOptimalConfiguration(nodes []*model.Node) *FlConfiguration {
	flGlobalAggregator := &model.FlAggregator{}
	flLocalAggregators := []*model.FlAggregator{}
	flClients := []*model.FlClient{}

	globalAggregator, localAggregators, clients := common.GetClientsAndAggregators(nodes)

	config.globalAgregatorNode = globalAggregator
	config.localAggregatorNodes = localAggregators

	if len(localAggregators) <= 1 {
		// centralized
		flGlobalAggregator = &model.FlAggregator{
			Id:              globalAggregator.Id,
			InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(common.GLOBAL_AGGREGATOR_PORT)),
			ExternalAddress: common.GetGlobalAggregatorExternalAddress(globalAggregator.Id),
			Port:            common.GLOBAL_AGGREGATOR_PORT,
			NumClients:      2, //int32(len(clients))
			Rounds:          common.GLOBAL_AGGREGATOR_ROUNDS,
		}
		flClients = common.ClientNodesToFlClients(clients, flGlobalAggregator, config.epochs*config.localRounds)

		return &FlConfiguration{
			GlobalAggregator: flGlobalAggregator,
			Clients:          flClients,
			Epochs:           config.epochs * config.localRounds,
		}
	}

	// note: this is simple example of clustering with equal distribution of clients per aggregator
	config.bestClusters = make([][]*model.Node, 0)

	// get cluster sizes
	numClients := len(clients)
	numLocalAggregators := len(localAggregators)
	div := numClients / numLocalAggregators
	mod := numClients % numLocalAggregators
	clusterSizes := make([]int, numLocalAggregators)
	for i := 0; i < numLocalAggregators; i++ {
		if i < mod {
			clusterSizes[i] = div + 1
		} else {
			clusterSizes[i] = div
		}
	}

	// make optimal clusters
	config.bestClusters = config.assignClientsGreedy(clients, localAggregators, clusterSizes)
	fmt.Print("Optimal clusters: ")
	printClusters(config.bestClusters)

	// prepare clients and aggregators
	flGlobalAggregator = &model.FlAggregator{
		Id:              globalAggregator.Id,
		InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(common.GLOBAL_AGGREGATOR_PORT)),
		ExternalAddress: common.GetGlobalAggregatorExternalAddress(globalAggregator.Id),
		Port:            common.GLOBAL_AGGREGATOR_PORT,
		NumClients:      int32(len(localAggregators)),
		Rounds:          common.GLOBAL_AGGREGATOR_ROUNDS,
	}
	for n, cluster := range config.bestClusters {
		localAggregator := localAggregators[n]
		localFlAggregator := &model.FlAggregator{
			Id:              localAggregator.Id,
			InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(common.LOCAL_AGGREGATOR_PORT)),
			ExternalAddress: common.GetLocalAggregatorExternalAddress(localAggregator.Id),
			Port:            common.LOCAL_AGGREGATOR_PORT,
			NumClients:      2, // int32(len(cluster))
			Rounds:          common.LOCAL_AGGREGATOR_ROUNDS,
			LocalRounds:     config.localRounds,
			ParentAddress:   flGlobalAggregator.ExternalAddress,
		}
		flLocalAggregators = append(flLocalAggregators, localFlAggregator)
		flClientsCluster := common.ClientNodesToFlClients(cluster, localFlAggregator, config.epochs)
		flClients = append(flClients, flClientsCluster...)
	}

	return &FlConfiguration{
		GlobalAggregator: flGlobalAggregator,
		LocalAggregators: flLocalAggregators,
		Clients:          flClients,
		Epochs:           config.epochs,
		LocalRounds:      config.localRounds,
	}
}

func (config *MinimizeCommCostGreedyConfiguration) assignClientsGreedy(clients []*model.Node, localAggregators []*model.Node, clusterSizes []int) [][]*model.Node {
	// Initialize clusters
	clusters := make([][]*model.Node, len(localAggregators))
	currentSizes := make([]int, len(localAggregators))

	for _, client := range clients {
		bestIdx := -1
		bestCost := float32(math.MaxFloat32)

		for i, la := range localAggregators {
			if currentSizes[i] >= clusterSizes[i] {
				continue // skip full cluster
			}
			cost := client.CommunicationCosts[la.Id]
			if cost < bestCost {
				bestCost = cost
				bestIdx = i
			}
		}

		if bestIdx != -1 {
			clusters[bestIdx] = append(clusters[bestIdx], client)
			currentSizes[bestIdx]++
		} else {
			// This should never happen if total clusterSizes == len(clients)
			panic("No available cluster for client " + client.Id)
		}
	}

	return clusters
}
