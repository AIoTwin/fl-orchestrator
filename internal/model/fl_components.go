package model

type FlClient struct {
	Id               string
	ParentAddress    string
	ParentNodeId     string
	Epochs           int32
	DataDistribution map[string]int64
}

type FlAggregator struct {
	Id              string
	InternalAddress string
	ExternalAddress string
	ParentAddress   string
	Port            int32
	NumClients      int32
	Rounds          int32
	LocalRounds     int32
}
