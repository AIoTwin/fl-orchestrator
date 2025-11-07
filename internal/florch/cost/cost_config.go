package cost

type CostConfiguration struct {
	CostType       string
	Budget         float32
	TargetAccuracy float32
}

const TotalBudget_CostType = "totalBudget"
const CostMinimization_CostType = "costMin"
