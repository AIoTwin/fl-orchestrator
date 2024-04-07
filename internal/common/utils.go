package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

func GetAvailableNodesFromFile() map[string]*model.Node {
	nodes := make(map[string]*model.Node)

	records := ReadCsvFile("../../configs/cluster/cluster.csv")
	for _, record := range records {
		communicationCosts := make(map[string]float32)
		commCostsSlice := strings.Split(record[3], ",")
		for _, commCost := range commCostsSlice {
			commCostSplited := strings.Split(commCost, ":")
			if len(commCostSplited) == 2 {
				costParsed, _ := strconv.ParseFloat(commCostSplited[1], 32)
				communicationCosts[commCostSplited[0]] = float32(costParsed)
			}
		}

		node := &model.Node{
			Id:                 record[0],
			InternalIp:         record[1],
			Resources:          model.NodeResources{},
			FlType:             record[2],
			CommunicationCosts: communicationCosts,
		}

		nodes[node.Id] = node
	}

	return nodes
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

func GetClientsAndAggregators(nodes []*model.Node) ([]*model.Node, []*model.Node) {
	clients := []*model.Node{}
	aggregators := []*model.Node{}
	for _, node := range nodes {
		if node.FlType == FL_TYPE_AGGREGATOR {
			aggregators = append(aggregators, node)
		} else {
			clients = append(clients, node)
		}
	}

	return clients, aggregators
}

func ClientNodesToFlClients(clients []*model.Node, flAggregator *model.FlAggregator, epochs int32) []*model.FlClient {
	flClients := []*model.FlClient{}
	for _, client := range clients {
		flClient := &model.FlClient{
			Id:            client.Id,
			ParentAddress: flAggregator.ExternalAddress,
			Epochs:        epochs,
		}

		flClients = append(flClients, flClient)
	}

	return flClients
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
