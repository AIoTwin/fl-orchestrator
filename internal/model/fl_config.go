package model

type ClientConfig struct {
	ClientId      string `yaml:"client_id"`
	ServerAddress string `yaml:"server_address"`
	Epochs        int32  `yaml:"epochs"`
}

type GlobalAggregatorConfig struct {
	ServerAddress string `yaml:"server_address"`
	Rounds        int32  `yaml:"rounds"`
}

type GlobalAggregatorEntryConfig struct {
	NumClients int32 `yaml:"num_clients"`
}

type LocalAggregatorConfig struct {
	ServerAddress string `yaml:"server_address"`
	ParentAddress string `yaml:"parent_address"`
	Rounds        int32  `yaml:"rounds"`
}

type LocalAggregatorEntryConfig struct {
	NumClients int32 `yaml:"num_clients"`
}
