package main

import (
	"os"
	"strconv"

	k8sorch "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch/k8s"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/cost"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/flconfig"
	"github.com/hashicorp/go-hclog"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "fl-orch",
		Level: hclog.LevelFromString("DEBUG"),
	})

	eventBus := events.NewEventBus()

	k8sOrchestrator, err := k8sorch.NewK8sOrchestrator("../../configs/cluster/kube_config.yaml", eventBus, false)
	if err != nil {
		logger.Error("Error while initializing k8s client ::", err.Error())
		return
	}

	modelSize, _ := strconv.ParseFloat(os.Args[1], 32)
	communicationBudget, _ := strconv.ParseFloat(os.Args[2], 32)

	costConfiguration := &cost.CostCofiguration{
		CostType:            cost.TotalBudget_CostType,
		CommunicationBudget: float32(communicationBudget),
	}

	flOrchestrator, err := florch.NewFlOrchestrator(k8sOrchestrator, eventBus, logger, flconfig.Cent_Hier_ConfigModelName,
		-1, -1, float32(modelSize), costConfiguration, false)
	if err != nil {
		logger.Error("Error creating orchestrator", "error", err)
		return
	}

	flOrchestrator.Start()
}
