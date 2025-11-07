package model

type FlEntities struct {
	GlobalAggregator *FlAggregator
	LocalAggregators []*FlAggregator
	Clients          []*FlClient
}

type ClientUtility struct {
	DatasetSizeScore      float32
	DataDistribution      []int
	DataDistributionScore float32
	ModelDifference       []float64
	ModelDifferenceScore  float32
	OverallUtility        float32
}
