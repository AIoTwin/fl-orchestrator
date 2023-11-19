package model

type FlClient struct {
	Id            string
	ParentAddress string
	Epochs        int32
}

type LocalAggregator struct {
}

type GlobalAggregator struct {
	Address    string
	Port       int32
	NumClients int32
	Rounds     int32
}
