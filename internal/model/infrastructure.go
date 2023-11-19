package model

type Node struct {
	Id         string
	InternalIp string
	Resources  *NodeResources
}

type NodeResources struct {
	CpuTotal float64
	RamTotal float64
	CpuUsage float64
	RamUsage float64
}
