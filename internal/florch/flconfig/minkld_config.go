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
	bestClusters        [][]*model.Node
	averageDistribution []float64
	bestKld             float64
}

func NewMinimizeKldConfiguration(epochs int32, localRounds int32) *MinimizeKldConfiguration {
	return &MinimizeKldConfiguration{
		epochs:      epochs,
		localRounds: localRounds,
	}
}

func (config *MinimizeKldConfiguration) GetOptimalConfiguration(nodes []*model.Node) *FlConfiguration {
	flGlobalAggregator := &model.FlAggregator{}
	flLocalAggregators := []*model.FlAggregator{}
	flClients := []*model.FlClient{}

	globalAggregator, localAggregators, clients := common.GetClientsAndAggregators(nodes)

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

func (config *MinimizeKldConfiguration) partitionClients(clients []*model.Node, index int, clusters [][]*model.Node, clusterSizes []int) {
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
func deepCopyClusters(clusters [][]*model.Node) [][]*model.Node {
	newClusters := make([][]*model.Node, len(clusters))
	for i := range clusters {
		newClusters[i] = make([]*model.Node, len(clusters[i]))
		copy(newClusters[i], clusters[i])
	}
	return newClusters
}
