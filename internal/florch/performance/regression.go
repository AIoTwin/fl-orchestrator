package performance

type Regression interface {
	PredictY(x float64) float64
	PredictX(y float64) float64
	PrintFunction() string
}
