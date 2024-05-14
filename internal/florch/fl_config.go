package florch

import "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"

type IFlConfiguration interface {
	GetOptimalConfiguration(nodes []*model.Node, modelSize float32, communicationBudget float32) ([]*model.FlClient, []*model.FlAggregator,
		int32, int32)
}
