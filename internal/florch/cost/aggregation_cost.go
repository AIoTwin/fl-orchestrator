package cost

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/flconfig"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

func HasID(ids []int, x int) bool {
	for _, id := range ids {
		if id == x {
			return true
		}
	}
	return false
}

func parseClientID(s string) (int, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if !strings.HasPrefix(s, "n") || len(s) == 1 {
		return 0, fmt.Errorf("bad id: %q (expected like n30)", s)
	}
	return strconv.Atoi(s[1:])
}

func GetGlobalRoundCost(configuration *flconfig.FlConfiguration, nodes map[string]*model.Node, modelSize float32, costSource CostSource, ids []int) float32 {

	if costSource == COMMUNICATION {
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
	//IF COST ENERGY THEN...

	gaCost := float32(0.0)
	gaNode := nodes[configuration.GlobalAggregator.Id]
	gaCost = gaNode.EnergyCost

	laCost := float32(0.0)
	for _, localAggregator := range configuration.LocalAggregators {
		laNode := nodes[localAggregator.Id]
		energyCost := laNode.EnergyCost * float32(configuration.LocalRounds)
		laCost += float32(energyCost)
	}

	clCost := float32(0.0)

	for _, client := range configuration.Clients {
		cl_id, _ := parseClientID(client.Id)
		if HasID(ids, cl_id) {
			continue
		}
		clNode := nodes[client.Id]
		energyCost := clNode.EnergyCost * float32(configuration.Epochs) * float32(configuration.LocalRounds)
		clCost += (energyCost)
	}

	globalRoundCost := gaCost + laCost + clCost

	return globalRoundCost

}

func GetReconfigurationChangeCost(oldConfiguration *flconfig.FlConfiguration, newConfiguration *flconfig.FlConfiguration,
	nodes map[string]*model.Node, modelSize float32, costSource CostSource) float32 {

	if costSource == ENERGY {
		return 0.0
	}

	//IF COST COMMUNICATION THEN...
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
