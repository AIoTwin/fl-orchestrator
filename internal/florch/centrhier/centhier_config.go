package centhier

import (
	"fmt"
	"math"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type CentrHierFlConfiguration struct {
}

func NewCentrHierFlConfiguration() *CentrHierFlConfiguration {
	return &CentrHierFlConfiguration{}
}

func (config *CentrHierFlConfiguration) GetOptimalConfiguration(nodes []*model.Node, modelSize float32, communicationBudget float32) ([]*model.FlClient, []*model.FlAggregator,
	int32, int32) {
	clientsC, aggregatorsC, epochsC := getOptimalConfigurationCentralized(nodes, modelSize, communicationBudget)
	clientsH, aaggregatorsH, epochsH, localRoundsH := getOptimalConfigurationHierarchical(nodes, modelSize, communicationBudget)

	if len(clientsH) == 0 {
		return clientsC, aggregatorsC, epochsC, 0
	}

	setup := getOptimalSetup(epochsC, epochsH, localRoundsH, 0)
	if setup == "centralized" {
		return clientsC, aggregatorsC, epochsC, 0
	} else {
		return clientsH, aaggregatorsH, epochsH, localRoundsH
	}
}

func getOptimalConfigurationCentralized(nodes []*model.Node, modelSize float32, communicationBudget float32) ([]*model.FlClient, []*model.FlAggregator,
	int32) {
	clients, aggregators := common.GetClientsAndAggregators(nodes)

	minEpochs := int32(math.MaxInt32)
	var bestAggregator *model.Node
	for _, aggregator := range aggregators {
		aggregationCost, err := calculateAggregationCost(clients, aggregator.Id, modelSize)
		if err != nil {
			continue
		}

		for n := 1; n < int(minEpochs); n++ {
			costPerEpoch := aggregationCost / float32(n)
			if costPerEpoch <= communicationBudget {
				minEpochs = int32(n)
				bestAggregator = aggregator
				break
			}
		}
	}

	flAggregator := &model.FlAggregator{
		Id:              bestAggregator.Id,
		InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(common.GLOBAL_AGGREGATOR_PORT)),
		ExternalAddress: common.GetGlobalAggregatorExternalAddress(bestAggregator.Id),
		Port:            common.GLOBAL_AGGREGATOR_PORT,
		NumClients:      int32(len(clients)),
		Rounds:          common.GLOBAL_AGGREGATOR_ROUNDS,
	}
	flClients := common.ClientNodesToFlClients(clients, flAggregator, int32(minEpochs))
	flAggregators := []*model.FlAggregator{
		flAggregator,
	}

	return flClients, flAggregators, minEpochs
}

func getOptimalConfigurationHierarchical(nodes []*model.Node, modelSize float32, communicationBudget float32) ([]*model.FlClient, []*model.FlAggregator,
	int32, int32) {
	epochs := int32(1)
	localRounds := int32(1)
	flAggregators := []*model.FlAggregator{}
	flClients := []*model.FlClient{}

	// note: this is dummy example of clustering with equal distribution of clients per aggregator
	clients, aggregators := common.GetClientsAndAggregators(nodes)
	globalAggregator := aggregators[0]
	localAggregators := aggregators[1:]

	if len(localAggregators) == 0 {
		return flClients, flAggregators, epochs, localRounds
	}

	div := len(clients) / len(localAggregators)
	mod := len(clients) % len(localAggregators)
	clusters := [][]*model.Node{}
	lastClientIndex := 0
	for i := 0; i < len(localAggregators); i++ {
		var cluster []*model.Node
		startIndex := lastClientIndex
		var endIndex int
		if i < mod {
			endIndex = startIndex + div + 1
		} else {
			endIndex = startIndex + div
		}
		cluster = clients[startIndex:endIndex]
		clusters = append(clusters, cluster)

		lastClientIndex = endIndex
	}

	globalFlAggregator := &model.FlAggregator{
		Id:              globalAggregator.Id,
		InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(common.GLOBAL_AGGREGATOR_PORT)),
		ExternalAddress: common.GetGlobalAggregatorExternalAddress(globalAggregator.Id),
		Port:            common.GLOBAL_AGGREGATOR_PORT,
		NumClients:      int32(len(aggregators) - 1),
		Rounds:          common.GLOBAL_AGGREGATOR_ROUNDS,
	}
	flAggregators = append(flAggregators, globalFlAggregator)
	for n, cluster := range clusters {
		localAggregator := aggregators[n+1]
		localFlAggregator := &model.FlAggregator{
			Id:              localAggregator.Id,
			InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(common.LOCAL_AGGREGATOR_PORT)),
			ExternalAddress: common.GetLocalAggregatorExternalAddress(localAggregator.Id),
			Port:            common.LOCAL_AGGREGATOR_PORT,
			NumClients:      int32(len(cluster)),
			Rounds:          common.LOCAL_AGGREGATOR_ROUNDS,
			LocalRounds:     localRounds,
			ParentAddress:   globalFlAggregator.ExternalAddress,
		}
		flAggregators = append(flAggregators, localFlAggregator)
		flClientsCluster := common.ClientNodesToFlClients(cluster, localFlAggregator, epochs)
		flClients = append(flClients, flClientsCluster...)
	}

	return flClients, flAggregators, epochs, localRounds
}

func getOptimalSetup(epochsCentralized int32, epochsHierarchical int32, localRoundsHierarchical int32, numClustersHierarchical int32) string {
	if epochsCentralized >= epochsHierarchical*localRoundsHierarchical {
		return "hierarachical"
	} else {
		return "centralized"
	}
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
