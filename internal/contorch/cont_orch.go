package contorch

import (
	"bytes"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type IContainerOrchestrator interface {
	GetAvailableNodes(initialRequest bool) (map[string]*model.Node, error)
	StartNodeStateChangeNotifier()
	StopAllNotifiers()
	CreateGlobalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error
	GetGlobalAggregatorLogs() (bytes.Buffer, error)
	RemoveGlobalAggregator(aggregator *model.FlAggregator) error
	CreateLocalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error
	RemoveLocalAggregator(aggregator *model.FlAggregator) error
	CreateFlClient(client *model.FlClient, configFiles map[string]string) error
	RemoveClient(client *model.FlClient) error
}
