package main

import (
	"os"
	"strconv"

	k8sorch "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch/k8s"
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

	k8sOrchestrator, err := k8sorch.NewK8sOrchestrator("/home/ubuntu/.kube/config", eventBus, false)
	if err != nil {
		logger.Error("Error while initializing k8s client ::", err.Error())
		return
	}

	centHierConfig := centhier.NewCentrHierFlConfiguration()

	modelSize, _ := strconv.ParseFloat(os.Args[1], 32)
	communicationBudget, _ := strconv.ParseFloat(os.Args[2], 32)
	flOrchestrator := florch.NewFlOrchestrator(k8sOrchestrator, centHierConfig, eventBus, logger, float32(modelSize), float32(communicationBudget))

	flOrchestrator.Start()
}
