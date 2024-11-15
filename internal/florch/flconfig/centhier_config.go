package flconfig

import (
	"fmt"
	"math"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type CentrHierFlConfiguration struct {
	modelSize           float32
	communicationBudget float32
	bestClusters        [][]*model.FlClient
	averageDistribution []float64
	bestKld             float64
}

func NewCentrHierFlConfiguration(modelSize float32, communicationBudget float32) *CentrHierFlConfiguration {
	return &CentrHierFlConfiguration{
		modelSize:           modelSize,
		communicationBudget: communicationBudget,
	}
}

func (config *CentrHierFlConfiguration) GetOptimalConfiguration(flEntitiesInitial *model.FlEntities) *FlConfiguration {
	var epochs int32
	var localRounds int32

	flEntities := copyFlEntitites(flEntitiesInitial)

	if len(flEntities.LocalAggregators) == 0 {
		flEntities, epochs = getOptimalConfigurationCentralized(flEntities, config.modelSize, config.communicationBudget)
	} else {
		flEntities, epochs, localRounds = config.getOptimalConfigurationHierarchical(flEntities, config.modelSize, config.communicationBudget)
	}

	return &FlConfiguration{
		FlEntities:  flEntities,
		Epochs:      epochs,
		LocalRounds: localRounds,
	}
}

func getOptimalConfigurationCentralized(flEntities *model.FlEntities, modelSize float32, communicationBudget float32) (*model.FlEntities, int32) {
	aggregationCost, err := calculateLocalAggregationCost(flEntities.Clients, flEntities.GlobalAggregator.Id, modelSize)
	if err != nil {
		return nil, 0
	}

	minEpochs := int32(1)
	for n := 1; n < math.MaxInt32; n++ {
		costPerEpoch := aggregationCost / float32(n)
		if costPerEpoch <= communicationBudget {
			minEpochs = int32(n)
			break
		}
	}

	flEntities.GlobalAggregator.NumClients = int32(len(flEntities.Clients))
	for _, client := range flEntities.Clients {
		client.Epochs = minEpochs
		client.ParentAddress = flEntities.GlobalAggregator.ExternalAddress
		client.ParentNodeId = flEntities.GlobalAggregator.Id
	}

	return flEntities, minEpochs
}

func (config *CentrHierFlConfiguration) getOptimalConfigurationHierarchical(flEntities *model.FlEntities, modelSize float32, communicationBudget float32) (
	*model.FlEntities, int32, int32) {
	epochs := int32(1)
	localRounds := int32(1)

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

	// optimize aggregation frequency within comm budget
	globalAggregationCost, localAggregationCost, _ := getHierarchicalAggregationCosts(flEntities.GlobalAggregator, flEntities.LocalAggregators,
		config.bestClusters, modelSize)
	costPerEpoch := globalAggregationCost + localAggregationCost
	if costPerEpoch > communicationBudget {
		for i := 0; i < math.MaxInt32; i++ {
			localRounds += 1
			costPerEpoch = (globalAggregationCost + float32(localRounds)*localAggregationCost) / (float32(epochs) * float32(localRounds))
			if costPerEpoch <= communicationBudget {
				break
			}

			for j := 0; j < 5; j++ {
				epochs += 1
				costPerEpoch = (globalAggregationCost + float32(localRounds)*localAggregationCost) / (float32(epochs) * float32(localRounds))
				if costPerEpoch <= communicationBudget {
					break
				}
			}
		}
	}

	fmt.Println("Cost per epoch:", costPerEpoch)

	// prepare clients and aggregators
	flEntities.GlobalAggregator.NumClients = int32(len(flEntities.LocalAggregators))
	clients := []*model.FlClient{}
	for n, cluster := range config.bestClusters {
		localAggregator := flEntities.LocalAggregators[n]
		localAggregator.LocalRounds = localRounds
		localAggregator.ParentAddress = flEntities.GlobalAggregator.ExternalAddress
		flClientsCluster := common.PrepareFlClients(cluster, localAggregator, epochs)
		clients = append(clients, flClientsCluster...)
	}
	flEntities.Clients = clients

	return flEntities, epochs, localRounds
}
