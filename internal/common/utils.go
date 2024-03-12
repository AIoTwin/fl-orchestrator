package common

import (
	"fmt"
	"strconv"
	"strings"

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

func GetAggregatorServiceName(aggregatorId string) string {
	return fmt.Sprintf("%s-%s", GLOBAL_AGGREGATOR_SERVICE_NAME, aggregatorId)
}

func GetAggregatorExternalAddress(aggregatorId string) string {
	return fmt.Sprintf("%s:%s", GetAggregatorServiceName(aggregatorId), fmt.Sprint(GLOBAL_AGGREGATOR_PORT))
}

func GetAggregatorConfigMapName(aggregatorId string) string {
	return fmt.Sprintf("%s-%s", GLOBAL_AGGREGATOR_CONFIG_MAP_NAME, aggregatorId)
}

func GetClientConfigMapName(clientId string) string {
	return fmt.Sprintf("%s-%s", FL_CLIENT_CONFIG_MAP_NAME, clientId)
}

func GetClientDeploymentName(clientId string) string {
	return fmt.Sprintf("%s-%s", FL_CLIENT_DEPLOYMENT_PREFIX, clientId)
}
