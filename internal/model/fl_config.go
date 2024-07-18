package model

type ClientEntryConfig struct {
	ClientId      int32  `yaml:"client_id"`
	Epochs        int32  `yaml:"epochs"`
	ServerAddress string `yaml:"server_address"`
}

type GlobalAggregatorEntryConfig struct {
	NumClients    int32  `yaml:"num_clients"`
	Rounds        int32  `yaml:"rounds"`
	ServerAddress string `yaml:"server_address"`
}

type LocalAggregatorEntryConfig struct {
	NumClients    int32  `yaml:"num_clients"`
	Rounds        int32  `yaml:"rounds"`
	LocalRounds   int32  `yaml:"local_rounds"`
	ParentAddress string `yaml:"parent_address"`
	ServerAddress string `yaml:"server_address"`
}

type LoggingConfig struct {
	RunName string `yaml:"run_name"`
}
