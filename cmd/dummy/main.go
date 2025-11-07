package main

import (
	"os"
	"strconv"

	dummyorch "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch/dummy"
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

	dummyOrchestrator := dummyorch.NewDummyOrch(eventBus)

	modelSize, _ := strconv.ParseFloat(os.Args[1], 32)
	communicationBudget, _ := strconv.ParseFloat(os.Args[2], 32)

	costConfiguration := &cost.CostConfiguration{
		CostType: cost.TotalBudget_CostType,
		Budget:   float32(communicationBudget),
	}

	flOrchestrator, err := florch.NewFlOrchestrator(dummyOrchestrator, eventBus, logger, flconfig.Cent_Hier_ConfigModelName,
		-1, -1, 32, 0.1, float32(modelSize), cost.COMMUNICATION, costConfiguration, false)
	if err != nil {
		logger.Error("Error creating orchestrator", "error", err)
		return
	}

	flOrchestrator.Start()
}
