package cost

import (
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/flconfig"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

func GetGlobalRoundCost(configuration *flconfig.FlConfiguration, nodes map[string]*model.Node, modelSize float32) float32 {
	gaCost := float32(0.0)
	for _, localAggregator := range configuration.LocalAggregators {
		laNode := nodes[localAggregator.Id]
		linkCost := laNode.CommunicationCosts[configuration.GlobalAggregator.Id]

		gaCost += linkCost * modelSize
	}

	laCost := float32(0.0)
	for _, client := range configuration.Clients {
		clientNode := nodes[client.Id]
		linkCost := clientNode.CommunicationCosts[client.ParentNodeId]

		if configuration.LocalRounds == 0 {
			laCost += linkCost * modelSize
		} else {
			laCost += float32(configuration.LocalRounds) * linkCost * modelSize
		}
	}

	globalRoundCost := gaCost + laCost
	return globalRoundCost
}

func GetReconfigurationChangeCost(oldConfiguration *flconfig.FlConfiguration, newConfiguration *flconfig.FlConfiguration,
	nodes map[string]*model.Node, modelSize float32) float32 {
	reconfigurationChangeCost := float32(0.0)

	for _, newClient := range newConfiguration.Clients {
		oldClient := common.GetClientInArray(oldConfiguration.Clients, newClient.Id)
		if (oldClient == &model.FlClient{} || newClient.ParentNodeId != oldClient.ParentNodeId) {
			newClientNode := nodes[newClient.Id]
			linkCost := newClientNode.CommunicationCosts[newClient.ParentNodeId]

			reconfigurationChangeCost += (linkCost / 2) * modelSize

			// add cost of downloading container image
		}
	}

	return reconfigurationChangeCost
}
