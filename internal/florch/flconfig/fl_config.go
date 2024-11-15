package flconfig

import "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"

type FlConfiguration struct {
	FlEntities  *model.FlEntities
	Epochs      int32
	LocalRounds int32
}

type IFlConfigurationModel interface {
	GetOptimalConfiguration(flEntitiesInitial *model.FlEntities) *FlConfiguration
}

const Cent_Hier_ConfigModelName = "centHier"
const MinimizeKld_ConfigModelName = "minKld"
const MinimizeCommCost_ConfigModelName = "minCommCost"
