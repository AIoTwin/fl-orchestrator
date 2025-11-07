package common

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

func GetAvailableNodesFromFile() (map[string]*model.Node, error) {
	nodes := make(map[string]*model.Node)

	records := ReadCsvFile("../../configs/cluster/cluster.csv")
	for _, record := range records {
		if len(record) != 6 {
			return nil, fmt.Errorf("Incorrect CSV record: %v", record)
		}

		communicationCosts := make(map[string]float32)
		commCostsSlice := strings.Split(record[2], ",")
		for _, commCost := range commCostsSlice {
			commCostSplited := strings.Split(commCost, ":")
			if len(commCostSplited) == 2 {
				costParsed, _ := strconv.ParseFloat(commCostSplited[1], 32)
				communicationCosts[commCostSplited[0]] = float32(costParsed)
			}
		}

		dataDistributions := make(map[string]int64)
		/* dataDistributionsSlice := strings.Split(record[3], ",")
		for _, dataDistribution := range dataDistributionsSlice {
			dataDistributionSplited := strings.Split(dataDistribution, ":")
			if len(dataDistributionSplited) == 2 {
				samplesParsed, _ := strconv.Atoi(dataDistributionSplited[1])
				dataDistributions[dataDistributionSplited[0]] = int64(samplesParsed)
			}
		} */

		energyCost, _ := strconv.ParseFloat(record[3], 32)
		numPartitions, _ := strconv.Atoi(record[4])
		partitionId, _ := strconv.Atoi(record[5])

		node := &model.Node{
			Id:                 record[0],
			Resources:          model.NodeResources{},
			FlType:             record[1],
			CommunicationCosts: communicationCosts,
			EnergyCost:         float32(energyCost),
			DataDistribution:   dataDistributions,
			NumPartitions:      int32(numPartitions),
			PartitionId:        int32(partitionId),
		}

		nodes[node.Id] = node
	}

	return nodes, nil
}

func GetNodeStateChangeEvent(availableNodesCurrent map[string]*model.Node, availableNodesNew map[string]*model.Node) events.Event {
	nodesAdded := []*model.Node{}
	// check for added nodes
	for _, node := range availableNodesNew {
		_, found := availableNodesCurrent[node.Id]
		if !found {
			nodesAdded = append(nodesAdded, node)
		}
	}

	nodesRemoved := []*model.Node{}
	// check for removed nodes
	for _, node := range availableNodesCurrent {
		_, found := availableNodesNew[node.Id]
		if !found {
			nodesRemoved = append(nodesRemoved, node)
		}
	}

	var event events.Event
	if len(nodesAdded) > 0 || len(nodesRemoved) > 0 {
		event = events.Event{
			Type:      NODE_STATE_CHANGE_EVENT_TYPE,
			Timestamp: time.Now(),
			Data: events.NodeStateChangeEvent{
				NodesAdded:   nodesAdded,
				NodesRemoved: nodesRemoved,
			},
		}
	}

	return event
}

func GetClientsAndAggregators(nodes []*model.Node) (*model.Node, []*model.Node, []*model.Node) {
	clients := []*model.Node{}
	localAggregators := []*model.Node{}
	globalAggregator := &model.Node{}
	for _, node := range nodes {
		switch node.FlType {
		case FL_TYPE_GLOBAL_AGGREGATOR:
			globalAggregator = node
		case FL_TYPE_LOCAL_AGGREGATOR:
			localAggregators = append(localAggregators, node)
		case FL_TYPE_CLIENT:
			clients = append(clients, node)
		}
	}

	sort.Slice(localAggregators, func(i, j int) bool {
		compare := strings.Compare(localAggregators[i].Id, localAggregators[j].Id)
		if compare == -1 {
			return true
		} else {
			return false
		}
	})

	return globalAggregator, localAggregators, clients
}

func ClientNodesToFlClients(clients []*model.Node, flAggregator *model.FlAggregator, epochs int32) []*model.FlClient {
	flClients := []*model.FlClient{}
	for _, client := range clients {
		flClient := &model.FlClient{
			Id:               client.Id,
			ParentAddress:    flAggregator.ExternalAddress,
			ParentNodeId:     flAggregator.Id,
			Epochs:           epochs,
			DataDistribution: client.DataDistribution,
			NumPartitions:    client.NumPartitions,
			PartitionId:      client.PartitionId,
		}

		flClients = append(flClients, flClient)
	}

	return flClients
}

func GetClientInArray(clients []*model.FlClient, clientId string) *model.FlClient {
	for _, client := range clients {
		if client.Id == clientId {
			return client
		}
	}

	return &model.FlClient{}
}

func GetGlobalAggregatorServiceName(aggregatorId string) string {
	return fmt.Sprintf("%s-%s", GLOBAL_AGGREGATOR_SERVICE_NAME, aggregatorId)
}

func GetGlobalAggregatorExternalAddress(aggregatorId string) string {
	return fmt.Sprintf("%s:%s", GetGlobalAggregatorServiceName(aggregatorId), fmt.Sprint(GLOBAL_AGGREGATOR_PORT))
}

func GetGlobalAggregatorConfigMapName(aggregatorId string) string {
	return fmt.Sprintf("%s-%s", GLOBAL_AGGREGATOR_CONFIG_MAP_NAME, aggregatorId)
}

func GetLocalAggregatorServiceName(aggregatorId string) string {
	return fmt.Sprintf("%s-%s", LOCAL_AGGREGATOR_SERVICE_NAME, aggregatorId)
}

func GetLocalAggregatorExternalAddress(aggregatorId string) string {
	return fmt.Sprintf("%s:%s", GetLocalAggregatorServiceName(aggregatorId), fmt.Sprint(LOCAL_AGGREGATOR_PORT))
}

func GetLocalAggregatorConfigMapName(aggregatorId string) string {
	return fmt.Sprintf("%s-%s", LOCAL_AGGREGATOR_CONFIG_MAP_NAME, aggregatorId)
}

func GetLocalAggregatorDeploymentName(aggregatorId string) string {
	return fmt.Sprintf("%s-%s", LOCAL_AGGRETATOR_DEPLOYMENT_PREFIX, aggregatorId)
}

func GetClientConfigMapName(clientId string) string {
	return fmt.Sprintf("%s-%s", FL_CLIENT_CONFIG_MAP_NAME, clientId)
}

func GetClientDeploymentName(clientId string) string {
	return fmt.Sprintf("%s-%s", FL_CLIENT_DEPLOYMENT_PREFIX, clientId)
}

func CalculateAverageFloat64(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	var sum float64
	for _, number := range numbers {
		sum += number
	}

	return sum / float64(len(numbers))
}
