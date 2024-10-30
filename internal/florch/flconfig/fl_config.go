package flconfig

import "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"

type FlConfiguration struct {
	GlobalAggregator *model.FlAggregator
	LocalAggregators []*model.FlAggregator
	Clients          []*model.FlClient
	Epochs           int32
	LocalRounds      int32
}

type IFlConfigurationModel interface {
	GetOptimalConfiguration(nodes []*model.Node) *FlConfiguration
}

const Cent_Hier_ConfigModelName = "centHier"
const MinimizeKld_ConfigModelName = "minKld"
const MinimizeCommCost_ConfigModelName = "minCommCost"
