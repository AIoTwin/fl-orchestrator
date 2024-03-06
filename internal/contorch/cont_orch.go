package contorch

import (
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type IContainerOrchestrator interface {
	GetAvailableNodes() ([]*model.Node, error)
	CreateGlobalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error
	CreateFlClient(client *model.FlClient, configFiles map[string]string) error
}
