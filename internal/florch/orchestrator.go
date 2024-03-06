package florch

import (
	"fmt"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"github.com/hashicorp/go-hclog"
)

type FlOrchestrator struct {
	contOrch contorch.IContainerOrchestrator
	flConfig IFlConfiguration
	logger   hclog.Logger
}

func NewFlOrchestrator(contOrch contorch.IContainerOrchestrator, flConfig IFlConfiguration, logger hclog.Logger) *FlOrchestrator {
	return &FlOrchestrator{
		contOrch: contOrch,
		flConfig: flConfig,
		logger:   logger,
	}
}

func (orch *FlOrchestrator) DeployAndStartFl(modelSize float32, communicationBudget float32) error {
	nodes, err := orch.contOrch.GetAvailableNodes()
	if err != nil {
		orch.logger.Error(err.Error())
		return err
	}

	clients, aggregators, epochs, localRounds := orch.flConfig.GetOptimalConfiguration(nodes, modelSize, communicationBudget)

	fmt.Println("Clients ::")
	for _, c := range clients {
		fmt.Printf("\t%+v\n", c)
	}
	fmt.Println("Aggregators ::")
	for _, a := range aggregators {
		fmt.Printf("\t%+v\n", a)
	}
	fmt.Println("Epochs: ", epochs)
	fmt.Println("Local rounds: ", localRounds)

	orch.DeployGlobalAggregator(aggregators[0])

	for _, client := range clients {
		orch.DeployFlClient(client)
	}

	return nil
}

func (orch *FlOrchestrator) DeployGlobalAggregator(flAggregator *model.FlAggregator) error {
	globalAggregatorConfigFilesData, err := BuildGlobalAggregatorConfigFiles(flAggregator)
	if err != nil {
		orch.logger.Error(fmt.Sprintf("Error while initializing global aggregator config files: %s", err.Error()))
		return err
	}

	if err := orch.contOrch.CreateGlobalAggregator(flAggregator, globalAggregatorConfigFilesData); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while deploying global aggregator: %s", err.Error()))
		return err
	}

	orch.logger.Info("Global aggregator deployed!")

	return nil
}

func (orch *FlOrchestrator) DeployFlClient(client *model.FlClient) error {
	clientConfigFilesData, err := BuildClientConfigFiles(client)
	if err != nil {
		orch.logger.Error(fmt.Sprintf("Error while initializing client %s config files: %s", client.Id, err.Error()))
		return err
	}

	if err := orch.contOrch.CreateFlClient(client, clientConfigFilesData); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while creating client %s deployment: %s", client.Id, err.Error()))
		return err
	}

	orch.logger.Info(fmt.Sprintf("Client %s deployed!", client.Id))

	return nil
}
