package florch

import (
	"fmt"
	"time"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"github.com/hashicorp/go-hclog"
)

type FlOrchestrator struct {
	contOrch            contorch.IContainerOrchestrator
	flConfig            IFlConfiguration
	eventBus            *events.EventBus
	logger              hclog.Logger
	modelSize           float32
	communicationBudget float32
	nodesMap            map[string]*model.Node
	globalAggregator    *model.FlAggregator
	localAggregators    []*model.FlAggregator
	clients             []*model.FlClient
}

func NewFlOrchestrator(contOrch contorch.IContainerOrchestrator, flConfig IFlConfiguration, eventBus *events.EventBus,
	logger hclog.Logger, modelSize float32, communicationBudget float32) *FlOrchestrator {
	return &FlOrchestrator{
		contOrch:            contOrch,
		flConfig:            flConfig,
		eventBus:            eventBus,
		logger:              logger,
		modelSize:           modelSize,
		communicationBudget: communicationBudget,
	}
}

func (orch *FlOrchestrator) Start() error {
	nodesMap, err := orch.contOrch.GetAvailableNodes(true)
	if err != nil {
		orch.logger.Error(err.Error())
		return err
	}
	orch.nodesMap = nodesMap

	orch.deployFl(false)

	nodeStateChangeChan := make(chan events.Event)
	orch.eventBus.Subscribe(common.NODE_STATE_CHANGE_EVENT_TYPE, nodeStateChangeChan)
	go orch.nodeStateChangeHandler(nodeStateChangeChan)

	go orch.contOrch.StartNodeStateChangeNotifier()

	flFinishedChan := make(chan events.Event)
	orch.eventBus.Subscribe(common.FL_FINISHED_EVENT_TYPE, flFinishedChan)
	orch.flFinishedHandler(flFinishedChan)

	return nil
}

func (orch *FlOrchestrator) deployFl(isReconfiguration bool) {
	nodesArray := nodesMapToArray(orch.nodesMap)

	globalAggregator, localAggregators, clients, epochs, localRounds := orch.flConfig.GetOptimalConfiguration(nodesArray, orch.modelSize, orch.communicationBudget)

	orch.globalAggregator = globalAggregator
	orch.localAggregators = localAggregators
	orch.clients = clients

	fmt.Println("Global aggregator ::")
	fmt.Printf("\t%+v\n", globalAggregator)
	fmt.Println("Local aggregators ::")
	for _, a := range localAggregators {
		fmt.Printf("\t%+v\n", a)
	}
	fmt.Println("Clients ::")
	for _, c := range clients {
		fmt.Printf("\t%+v\n", c)
	}
	fmt.Println("Epochs: ", epochs)
	fmt.Println("Local rounds: ", localRounds)

	if !isReconfiguration {
		orch.deployGlobalAggregator(globalAggregator)
		time.Sleep(10 * time.Second)
	}

	for _, localAggregator := range localAggregators {
		orch.deployLocalAggregator(localAggregator)
		time.Sleep(1 * time.Second)
	}
	time.Sleep(10 * time.Second)

	for _, client := range clients {
		orch.deployFlClient(client)
		time.Sleep(1 * time.Second)
	}
}

func (orch *FlOrchestrator) removeFl(isReconfiguration bool) {
	for _, client := range orch.clients {
		orch.contOrch.RemoveClient(client)
	}

	for _, localAggregator := range orch.localAggregators {
		orch.contOrch.RemoveLocalAggregator(localAggregator)
	}

	if !isReconfiguration {
		orch.contOrch.RemoveGlobalAggregator(orch.globalAggregator)
	}
}

func (orch *FlOrchestrator) deployGlobalAggregator(flAggregator *model.FlAggregator) error {
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

func (orch *FlOrchestrator) deployLocalAggregator(flAggregator *model.FlAggregator) error {
	localAggregatorConfigFilesData, err := BuildLocalAggregatorConfigFiles(flAggregator)
	if err != nil {
		orch.logger.Error(fmt.Sprintf("Error while initializing local aggregator config files: %s", err.Error()))
		return err
	}

	if err := orch.contOrch.CreateLocalAggregator(flAggregator, localAggregatorConfigFilesData); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while deploying local aggregator: %s", err.Error()))
		return err
	}

	orch.logger.Info("Local aggregator deployed!")

	return nil
}

func (orch *FlOrchestrator) deployFlClient(client *model.FlClient) error {
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

func (orch *FlOrchestrator) flFinishedHandler(eventChan <-chan events.Event) {
	for event := range eventChan {
		flFinishedEvent, ok := event.Data.(events.FlFinishedEvent)
		if !ok {
			fmt.Println("Invalid event data")
			continue
		}

		// Handle the event
		fmt.Println("New event:")
		fmt.Println("Timestamp:", event.Timestamp)
		fmt.Println("Exit code:", flFinishedEvent.ExitCode)
		fmt.Println("Exit message:", flFinishedEvent.ExitMessage)

		return
	}
}

func (orch *FlOrchestrator) nodeStateChangeHandler(eventChan <-chan events.Event) {
	for event := range eventChan {
		nodeStateChangeEvent, ok := event.Data.(events.NodeStateChangeEvent)
		if !ok {
			fmt.Println("Invalid event data")
			continue
		}

		// Handle the event
		fmt.Println("New event:")
		fmt.Println("Nodes added:", nodeStateChangeEvent.NodesAdded)
		fmt.Println("Node removed:", nodeStateChangeEvent.NodesRemoved)

		for _, node := range nodeStateChangeEvent.NodesAdded {
			orch.nodesMap[node.Id] = node
		}

		for _, node := range nodeStateChangeEvent.NodesRemoved {
			delete(orch.nodesMap, node.Id)
		}

		orch.removeFl(true)

		fmt.Println("Removed previous deployment and reconfiguring...")

		time.Sleep(5 * time.Second)

		orch.deployFl(true)
	}
}

func nodesMapToArray(nodesMap map[string]*model.Node) []*model.Node {
	nodesArray := make([]*model.Node, 0, len(nodesMap))

	for _, node := range nodesMap {
		nodesArray = append(nodesArray, node)
	}

	return nodesArray
}
