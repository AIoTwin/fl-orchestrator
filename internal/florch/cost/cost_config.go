package cost

type CostCofiguration struct {
	CostType            string
	CommunicationBudget float32
	TargetAccuracy      float32
}

const TotalBudget_CostType = "totalBudget"
const CostMinimization_CostType = "costMin"
