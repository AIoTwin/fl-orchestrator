package dummyorch

import (
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

type DummyOrch struct {
}

func NewDummyOrch() *DummyOrch {
	return &DummyOrch{}
}

func (dummyOrch *DummyOrch) GetAvailableNodes() ([]*model.Node, error) {
	nodes := fourNodesTestCluster()

	return nodes, nil
}

func (dummyOrch *DummyOrch) CreateGlobalAggregator(aggregator *model.FlAggregator, configFiles map[string]string) error {
	return nil
}

func (dummyOrch *DummyOrch) CreateFlClient(client *model.FlClient, configFiles map[string]string) error {
	return nil
}

func threeNodesTestCluster() []*model.Node {
	return []*model.Node{
		{
			Id:         "n1",
			InternalIp: "10.19.4.101",
			Resources:  model.NodeResources{},
			FlType:     "aggregator",
			CommunicationCosts: map[string]float32{
				"n2": 100,
				"n3": 100,
			},
		},
		{
			Id:         "n2",
			InternalIp: "10.19.4.102",
			Resources:  model.NodeResources{},
			FlType:     "client",
			CommunicationCosts: map[string]float32{
				"n1": 100,
				"n3": 10,
			},
		},
		{
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

func fourNodesTestCluster() []*model.Node {
	return []*model.Node{
		{
			Id:                 "n1",
			InternalIp:         "10.19.4.101",
			Resources:          model.NodeResources{},
			FlType:             "aggregator",
			CommunicationCosts: map[string]float32{},
		},
		{
			Id:                 "n2",
			InternalIp:         "10.19.4.102",
			Resources:          model.NodeResources{},
			FlType:             "aggregator",
			CommunicationCosts: map[string]float32{},
		},
		{
			Id:         "n3",
			InternalIp: "10.19.4.103",
			Resources:  model.NodeResources{},
			FlType:     "client",
			CommunicationCosts: map[string]float32{
				"n1": 100,
				"n2": 50,
			},
		},
		{
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
