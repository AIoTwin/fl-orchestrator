package main

import (
	"fmt"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	k8sorch "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch/k8s"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch"
	centhier "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/centrhier"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"github.com/hashicorp/go-hclog"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "fl-orch",
		Level: hclog.LevelFromString("DEBUG"),
	})

	/* dummyOrch := dummyorch.NewDummyOrch()
	centHierConfig := centhier.NewCentrHierFlConfiguration()

	flOrchestrator := florch.NewFlOrchestrator(dummyOrch, centHierConfig, logger)
	flOrchestrator.DeployAndStartFl(10, 600) */

	k8sOrchestrator, err := k8sorch.NewK8sOrchestrator("/home/ubuntu/.kube/config")
	if err != nil {
		logger.Error("Error while initializing k8s client ::", err.Error())
		return
	}

	centHierConfig := centhier.NewCentrHierFlConfiguration()

	flOrchestrator := florch.NewFlOrchestrator(k8sOrchestrator, centHierConfig, logger)

	flOrchestrator.DeployAndStartFl(10, 600)

	//deployHardcodedConfig(flOrchestrator)
}

func deployHardcodedConfig(flOrchestrator *florch.FlOrchestrator) {
	// global aggregator
	port := int32(30000)
	globalAggregator := &model.FlAggregator{
		Id:              "k3s-master",
		InternalAddress: fmt.Sprintf("%s:%s", "0.0.0.0", fmt.Sprint(port)),
		Port:            port,
		NumClients:      2,
		Rounds:          10,
	}

	flOrchestrator.DeployGlobalAggregator(globalAggregator)

	// clients
	globalAggregatorExternalAddress := fmt.Sprintf("%s:%s", common.GetAggregatorServiceName(globalAggregator.Id), fmt.Sprint(port))

	client1 := &model.FlClient{
		Id:            "survey-orch1",
		ParentAddress: globalAggregatorExternalAddress,
		Epochs:        2,
	}
	client2 := &model.FlClient{
		Id:            "fer-iot",
		ParentAddress: globalAggregatorExternalAddress,
		Epochs:        2,
	}

	flOrchestrator.DeployFlClient(client1)
	flOrchestrator.DeployFlClient(client2)
}
