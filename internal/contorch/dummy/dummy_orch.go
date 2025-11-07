package dummyorch

import (
	"bytes"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"github.com/robfig/cron/v3"
)

type DummyOrch struct {
	eventBus       *events.EventBus
	cronScheduler  *cron.Cron
	availableNodes map[string]*model.Node
}

func NewDummyOrch(eventBus *events.EventBus) *DummyOrch {
	return &DummyOrch{
		eventBus:       eventBus,
		cronScheduler:  cron.New(cron.WithSeconds()),
		availableNodes: make(map[string]*model.Node),
	}
}

func (orch *DummyOrch) GetAvailableNodes(initialRequest bool) (map[string]*model.Node, error) {
	nodes, err := common.GetAvailableNodesFromFile()
	if err != nil {
		return nil, err
	}

	if initialRequest {
		orch.availableNodes = nodes
	}

	return nodes, nil
}

func (orch *DummyOrch) StartNodeStateChangeNotifier() {
	orch.cronScheduler.AddFunc("@every 1s", orch.notifyNodeStateChanges)

	orch.cronScheduler.Start()
}

func (orch *DummyOrch) StopAllNotifiers() {
	orch.cronScheduler.Stop()
}

func (orch *DummyOrch) notifyNodeStateChanges() {
	availableNodesNew, err := orch.GetAvailableNodes(false)
	if err != nil {
		return
	}

	event := common.GetNodeStateChangeEvent(orch.availableNodes, availableNodesNew)
	if (event != events.Event{}) {
		orch.eventBus.Publish(event)
	}

	orch.availableNodes = availableNodesNew
}

func (orch *DummyOrch) CreateGlobalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error {
	return nil
}

func (orch *DummyOrch) GetGlobalAggregatorLogs() (bytes.Buffer, error) {
	return bytes.Buffer{}, nil
}

func (orch *DummyOrch) RemoveGlobalAggregator(aggregator *model.FlAggregator) error {
	return nil
}

func (orch *DummyOrch) CreateLocalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error {
	return nil
}

func (orch *DummyOrch) RemoveLocalAggregator(aggregator *model.FlAggregator) error {
	return nil
}

func (orch *DummyOrch) CreateFlClient(client *model.FlClient, configFiles map[string]string) error {
	return nil
}

func (orch *DummyOrch) RemoveClient(client *model.FlClient) error {
	return nil
}

func (orch *DummyOrch) GetClientLogs(clientsId string) (bytes.Buffer, error) {
	return bytes.Buffer{}, nil
}

func threeNodesTestCluster() map[string]*model.Node {
	return map[string]*model.Node{
		"n1": {
			Id:         "n1",
			InternalIp: "10.19.4.101",
			Resources:  model.NodeResources{},
			FlType:     "aggregator",
			CommunicationCosts: map[string]float32{
				"n2": 100,
				"n3": 100,
			},
		},
		"n2": {
			Id:         "n2",
			InternalIp: "10.19.4.102",
			Resources:  model.NodeResources{},
			FlType:     "client",
			CommunicationCosts: map[string]float32{
				"n1": 100,
				"n3": 10,
			},
		},
		"n3": {
			Id:         "n3",
			InternalIp: "10.19.4.103",
			Resources:  model.NodeResources{},
			FlType:     "client",
			CommunicationCosts: map[string]float32{
				"n1": 100,
				"n2": 10,
			},
		},
	}
}

func fourNodesTestCluster() map[string]*model.Node {
	return map[string]*model.Node{
		"n1": {
			Id:                 "n1",
			InternalIp:         "10.19.4.101",
			Resources:          model.NodeResources{},
			FlType:             "aggregator",
			CommunicationCosts: map[string]float32{},
		},
		"n2": {
			Id:                 "n2",
			InternalIp:         "10.19.4.102",
			Resources:          model.NodeResources{},
			FlType:             "aggregator",
			CommunicationCosts: map[string]float32{},
		},
		"n3": {
			Id:         "n3",
			InternalIp: "10.19.4.103",
			Resources:  model.NodeResources{},
			FlType:     "client",
			CommunicationCosts: map[string]float32{
				"n1": 100,
				"n2": 50,
			},
		},
		"n4": {
			Id:         "n4",
			InternalIp: "10.19.4.104",
			Resources:  model.NodeResources{},
			FlType:     "client",
			CommunicationCosts: map[string]float32{
				"n1": 100,
				"n2": 50,
			},
		},
	}
}
