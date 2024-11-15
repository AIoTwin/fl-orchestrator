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
		if len(record) != 4 {
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
		dataDistributionsSlice := strings.Split(record[3], ",")
		for _, dataDistribution := range dataDistributionsSlice {
			dataDistributionSplited := strings.Split(dataDistribution, ":")
			if len(dataDistributionSplited) == 2 {
				samplesParsed, _ := strconv.Atoi(dataDistributionSplited[1])
				dataDistributions[dataDistributionSplited[0]] = int64(samplesParsed)
			}
		}

		node := &model.Node{
			Id:                 record[0],
			Resources:          model.NodeResources{},
			FlType:             record[1],
			CommunicationCosts: communicationCosts,
			DataDistribution:   dataDistributions,
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

func GetClientsAndAggregators(nodes []*model.Node) (*model.FlAggregator, []*model.FlAggregator, []*model.FlClient) {
	globalAggregator := &model.FlAggregator{}
	localAggregators := []*model.FlAggregator{}
	clients := []*model.FlClient{}
	for _, node := range nodes {
		switch node.FlType {
		case FL_TYPE_GLOBAL_AGGREGATOR:
			globalAggregator = &model.FlAggregator{
				Id:              node.Id,
				InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(GLOBAL_AGGREGATOR_PORT)),
				ExternalAddress: GetGlobalAggregatorExternalAddress(node.Id),
				Port:            GLOBAL_AGGREGATOR_PORT,
				Rounds:          GLOBAL_AGGREGATOR_ROUNDS,
			}
		case FL_TYPE_LOCAL_AGGREGATOR:
			localAggregator := &model.FlAggregator{
				Id:                 node.Id,
				InternalAddress:    fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(LOCAL_AGGREGATOR_PORT)),
				ExternalAddress:    GetLocalAggregatorExternalAddress(node.Id),
				Port:               LOCAL_AGGREGATOR_PORT,
				NumClients:         2, // int32(len(cluster))
				Rounds:             LOCAL_AGGREGATOR_ROUNDS,
				CommunicationCosts: node.CommunicationCosts,
			}
			localAggregators = append(localAggregators, localAggregator)
		case FL_TYPE_CLIENT:
			client := &model.FlClient{
				Id:                 node.Id,
				CommunicationCosts: node.CommunicationCosts,
				DataDistribution:   node.DataDistribution,
			}
			clients = append(clients, client)
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

func PrepareFlClients(clients []*model.FlClient, flAggregator *model.FlAggregator, epochs int32) []*model.FlClient {
	for _, client := range clients {
		client.ParentAddress = flAggregator.ExternalAddress
		client.ParentNodeId = flAggregator.Id
		client.Epochs = epochs
	}

	return clients
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
