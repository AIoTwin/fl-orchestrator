package model

type Node struct {
	Id                 string
	InternalIp         string
	Resources          NodeResources
	FlType             string
	EnergyCost         float32
	CommunicationCosts map[string]float32 // destination node ID -> cost
	DataDistribution   map[string]int64   // class ID -> number of samples
	NumPartitions      int32
	PartitionId        int32
}

type NodeResources struct {
	CpuTotal float64
	RamTotal float64
	CpuUsage float64
	RamUsage float64
}
