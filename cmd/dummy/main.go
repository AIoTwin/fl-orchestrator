package main

import (
	"os"
	"strconv"

	dummyorch "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch/dummy"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch"
	centhier "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/centrhier"
	"github.com/hashicorp/go-hclog"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "fl-orch",
		Level: hclog.LevelFromString("DEBUG"),
	})

	eventBus := events.NewEventBus()

	dummyOrchestrator := dummyorch.NewDummyOrch(eventBus)

	centHierConfig := centhier.NewCentrHierFlConfiguration()

	modelSize, _ := strconv.ParseFloat(os.Args[1], 32)
	communicationBudget, _ := strconv.ParseFloat(os.Args[2], 32)
	flOrchestrator := florch.NewFlOrchestrator(dummyOrchestrator, centHierConfig, eventBus, logger, float32(modelSize), float32(communicationBudget))

	flOrchestrator.Start()
}
