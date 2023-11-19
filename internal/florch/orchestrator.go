package florch

import (
	"fmt"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"github.com/hashicorp/go-hclog"
)

type FlOrchestrator struct {
	contOrch contorch.IContainerOrchestrator
	logger   hclog.Logger
}

func NewFlOrchestrator(contOrch contorch.IContainerOrchestrator, logger hclog.Logger) *FlOrchestrator {
	return &FlOrchestrator{
		contOrch: contOrch,
		logger:   logger,
	}
}

func (orch *FlOrchestrator) DeployGlobalAggregator(globalAggregator model.GlobalAggregator) error {
	globalAggregatorConfigFilesData, err := BuildGlobalAggregatorConfigFiles(globalAggregator)
	if err != nil {
		orch.logger.Error(fmt.Sprintf("Error while initializing global aggregator config files: %s", err.Error()))
		return err
	}
	orch.contOrch.CreateConfigMapFromFiles("gacm", globalAggregatorConfigFilesData)

	globalAggregatorDeployment := BuildGlobalAggregatorDeployment(globalAggregator)
	if err := orch.contOrch.CreateDeployment(globalAggregatorDeployment); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while deploying global aggregator: %s", err.Error()))
		return err
	}

	globalAggregatorService := BuildGlobalAggregatorService(globalAggregator)
	if err := orch.contOrch.CreateService(globalAggregatorService); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while creating global aggregator service: %s", err.Error()))
		return err
	}

	orch.logger.Info("Global aggregator deployed!")

	return nil
}

func (orch *FlOrchestrator) DeployFlClient(client model.FlClient) error {
	clientConfigFilesData, err := BuildClientConfigFiles(client)
	if err != nil {
		orch.logger.Error(fmt.Sprintf("Error while initializing client %s config files: %s", client.Id, err.Error()))
		return err
	}
	orch.contOrch.CreateConfigMapFromFiles(fmt.Sprintf("clientcm-%s", client.Id), clientConfigFilesData)

	clientDeployment := BuildClientDeployment(client)
	if err := orch.contOrch.CreateDeployment(clientDeployment); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while creating client %s deployment: %s", client.Id, err.Error()))
		return err
	}

	orch.logger.Info(fmt.Sprintf("Client %s deployed!", client.Id))

	return nil
}
