package main

import (
	"fmt"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"github.com/hashicorp/go-hclog"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "fl-orch",
		Level: hclog.LevelFromString("DEBUG"),
	})

	k8sOrchestrator, err := contorch.NewK8sOrchestrator("/home/ivan/.kube/config")
	if err != nil {
		logger.Error("Error while initializing k8s client ::", err.Error())
		return
	}

	flOrchestrator := florch.NewFlOrchestrator(k8sOrchestrator, logger)

	// global aggregator
	port := int32(30000)
	globalAggregator := model.GlobalAggregator{
		Address:    fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(port)),
		Port:       port,
		NumClients: 2,
		Rounds:     10,
	}

	flOrchestrator.DeployGlobalAggregator(globalAggregator)

	// clients
	globalAggregatorExternalAddress := fmt.Sprintf("%s:%s", florch.GLOBAL_AGGREGATOR_SERVICE_NAME, fmt.Sprint(port))

	client1 := model.FlClient{
		Id:            "1",
		ParentAddress: globalAggregatorExternalAddress,
		Epochs:        2,
	}
	client2 := model.FlClient{
		Id:            "2",
		ParentAddress: globalAggregatorExternalAddress,
		Epochs:        2,
	}

	flOrchestrator.DeployFlClient(client1)
	flOrchestrator.DeployFlClient(client2)
}
