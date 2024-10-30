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
	bestClusters        [][]*model.Node
	averageDistribution []float64
	bestKld             float64
}

func NewCentrHierFlConfiguration(modelSize float32, communicationBudget float32) *CentrHierFlConfiguration {
	return &CentrHierFlConfiguration{
		modelSize:           modelSize,
		communicationBudget: communicationBudget,
	}
}

func (config *CentrHierFlConfiguration) GetOptimalConfiguration(nodes []*model.Node) *FlConfiguration {
	var globalAggregator *model.FlAggregator
	var localAggregators []*model.FlAggregator
	var clients []*model.FlClient
	var epochs int32
	var localRounds int32

	_, potentialLocalAggregators, _ := common.GetClientsAndAggregators(nodes)
	if len(potentialLocalAggregators) == 0 {
		globalAggregator, clients, epochs = getOptimalConfigurationCentralized(nodes, config.modelSize, config.communicationBudget)
	} else {
		globalAggregator, localAggregators, clients, epochs, localRounds = config.getOptimalConfigurationHierarchical(nodes,
			config.modelSize, config.communicationBudget)
	}

	return &FlConfiguration{
		GlobalAggregator: globalAggregator,
		LocalAggregators: localAggregators,
		Clients:          clients,
		Epochs:           epochs,
		LocalRounds:      localRounds,
	}
}

func getOptimalConfigurationCentralized(nodes []*model.Node, modelSize float32, communicationBudget float32) (*model.FlAggregator, []*model.FlClient,
	int32) {
	globalAggregator, _, clients := common.GetClientsAndAggregators(nodes)

	aggregationCost, err := calculateAggregationCost(clients, globalAggregator.Id, modelSize)
	if err != nil {
		return nil, nil, 0
	}

	minEpochs := int32(1)
	for n := 1; n < math.MaxInt32; n++ {
		costPerEpoch := aggregationCost / float32(n)
		if costPerEpoch <= communicationBudget {
			minEpochs = int32(n)
			break
		}
	}

	flGlobalAggregator := &model.FlAggregator{
		Id:              globalAggregator.Id,
		InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(common.GLOBAL_AGGREGATOR_PORT)),
		ExternalAddress: common.GetGlobalAggregatorExternalAddress(globalAggregator.Id),
		Port:            common.GLOBAL_AGGREGATOR_PORT,
		NumClients:      int32(len(clients)),
		Rounds:          common.GLOBAL_AGGREGATOR_ROUNDS,
	}
	flClients := common.ClientNodesToFlClients(clients, flGlobalAggregator, int32(minEpochs))

	return flGlobalAggregator, flClients, minEpochs
}

func (config *CentrHierFlConfiguration) getOptimalConfigurationHierarchical(nodes []*model.Node, modelSize float32, communicationBudget float32) (
	*model.FlAggregator, []*model.FlAggregator, []*model.FlClient, int32, int32) {
	epochs := int32(1)
	localRounds := int32(1)
	flGlobalAggregator := &model.FlAggregator{}
	flLocalAggregators := []*model.FlAggregator{}
	flClients := []*model.FlClient{}

	// note: this is dummy example of clustering with equal distribution of clients per aggregator
	globalAggregator, localAggregators, clients := common.GetClientsAndAggregators(nodes)

	config.bestClusters = make([][]*model.Node, 0)
	config.averageDistribution = make([]float64, 0)
	config.bestKld = math.MaxFloat64

	// get cluster sizes
	numClients := len(clients)
	numLocalAggregators := len(localAggregators)
	div := numClients / numLocalAggregators
	mod := numClients % numLocalAggregators
	clusters := make([][]*model.Node, numLocalAggregators)
	clusterSizes := make([]int, numLocalAggregators)
	for i := 0; i < numLocalAggregators; i++ {
		if i < mod {
			clusterSizes[i] = div + 1
		} else {
			clusterSizes[i] = div
		}
	}

	// make optimal clusters
	config.averageDistribution = getClusterDataDistribution(clients)
	config.bestKld = math.MaxFloat64
	config.partitionClients(clients, 0, clusters, clusterSizes)
	fmt.Print("Optimal clusters: ")
	printClusters(config.bestClusters)
	fmt.Println("Best KLD: ", config.bestKld)

	// optimize aggregation frequency within comm budget
	globalAggregationCost, localAggregationCost, _ := getHierarchicalAggregationCosts(globalAggregator, localAggregators, config.bestClusters, modelSize)
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
			LocalRounds:     localRounds,
			ParentAddress:   flGlobalAggregator.ExternalAddress,
		}
		flLocalAggregators = append(flLocalAggregators, localFlAggregator)
		flClientsCluster := common.ClientNodesToFlClients(cluster, localFlAggregator, epochs)
		flClients = append(flClients, flClientsCluster...)
	}

	return flGlobalAggregator, flLocalAggregators, flClients, epochs, localRounds
}
