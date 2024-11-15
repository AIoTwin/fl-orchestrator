package model

type FlEntities struct {
	GlobalAggregator *FlAggregator
	LocalAggregators []*FlAggregator
	Clients          []*FlClient
}

type FlClient struct {
	Id                 string
	ParentAddress      string
	ParentNodeId       string
	Epochs             int32
	CommunicationCosts map[string]float32
	DataDistribution   map[string]int64
	ClientUtility      ClientUtility
}

type ClientUtility struct {
	DatasetSizeScore      float32
	DataDistributionScore float32
	ModelDifference       []float64
	ModelDifferenceScore  float32
	OverallUtility        float32
}

type FlAggregator struct {
	Id                 string
	InternalAddress    string
	ExternalAddress    string
	ParentAddress      string
	Port               int32
	NumClients         int32
	Rounds             int32
	LocalRounds        int32
	CommunicationCosts map[string]float32
}
