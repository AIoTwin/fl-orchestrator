package centhier

import (
	"fmt"
	"math"
	"strconv"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type CentrHierFlConfiguration struct {
	CurrentClusters     [][]*model.Node
	BestClusters        [][]*model.Node
	AverageDistribution []float64
	BestKld             float64
}

func NewCentrHierFlConfiguration() *CentrHierFlConfiguration {
	return &CentrHierFlConfiguration{}
}

func (config *CentrHierFlConfiguration) GetOptimalConfiguration(nodes []*model.Node, modelSize float32, communicationBudget float32) (*model.FlAggregator,
	[]*model.FlAggregator, []*model.FlClient, int32, int32) {
	var globalAggregator *model.FlAggregator
	var localAggregators []*model.FlAggregator
	var clients []*model.FlClient
	var epochs int32
	var localRounds int32

	_, potentialLocalAggregators, _ := common.GetClientsAndAggregators(nodes)
	if len(potentialLocalAggregators) == 0 {
		globalAggregator, clients, epochs = getOptimalConfigurationCentralized(nodes, modelSize, communicationBudget)
	} else {
		globalAggregator, localAggregators, clients, epochs, localRounds = config.getOptimalConfigurationHierarchical(nodes, modelSize, communicationBudget)
	}

	return globalAggregator, localAggregators, clients, epochs, localRounds
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

	config.BestClusters = make([][]*model.Node, 0)
	config.AverageDistribution = make([]float64, 0)
	config.BestKld = math.MaxFloat64

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
	config.AverageDistribution = getDataDistribution(clients)
	config.BestKld = math.MaxFloat64
	config.partitionClients(clients, 0, clusters, clusterSizes)
	fmt.Print("Optimal clusters: ")
	printClusters(config.BestClusters)
	fmt.Println("Best KLD: ", config.BestKld)

	// optimize aggregation frequency within comm budget
	globalAggregationCost, localAggregationCost, _ := getHierarchicalAggregationCosts(globalAggregator, localAggregators, config.BestClusters, modelSize)
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
	for n, cluster := range config.BestClusters {
		localAggregator := localAggregators[n]
		localFlAggregator := &model.FlAggregator{
			Id:              localAggregator.Id,
			InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(common.LOCAL_AGGREGATOR_PORT)),
			ExternalAddress: common.GetLocalAggregatorExternalAddress(localAggregator.Id),
			Port:            common.LOCAL_AGGREGATOR_PORT,
			NumClients:      int32(len(cluster)),
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
			kld := getTotalKld(clusters, config.AverageDistribution)
			if kld < config.BestKld {
				config.BestKld = kld
				config.BestClusters = make([][]*model.Node, len(clusters))
				copy(config.BestClusters, clusters)
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
		clusterDataDistribution := getDataDistribution(cluster)
		klds[i] = klDivergence(clusterDataDistribution, averageDistribution)
	}

	return calculateAverage(klds)
}

func getDataDistribution(nodes []*model.Node) []float64 {
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

// Calculate Kullback-Leibler divergence between two distributions
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
