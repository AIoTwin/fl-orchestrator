package performance

import (
	"math"
)

const LogarithmicRegression_PredictionType = "log-reg"

type PerformancePrediction struct {
	regressionFunctionAccuracies Regression
	regressionFunctionLosses     Regression
}

func NewPerformancePrediction(accuracies []float32, losses []float32, predictionType string) *PerformancePrediction {
	pp := &PerformancePrediction{}

	accXs, accYs := prepareXAndY(accuracies)
	lossXs, lossYs := prepareXAndY(losses)

	if predictionType == LogarithmicRegression_PredictionType {
		pp.regressionFunctionAccuracies = NewLogarithmicRegression(accXs, accYs)
		pp.regressionFunctionLosses = NewLogarithmicRegression(lossXs, lossYs)
	}

	return pp
}

func (pp *PerformancePrediction) PredictAccuracy(round int32) float32 {
	return float32(pp.regressionFunctionAccuracies.PredictY(float64(round)))
}

func (pp *PerformancePrediction) PredictRoundForAccuracy(accuracy float32) int32 {
	predictedRoundFloat64 := math.Ceil(pp.regressionFunctionAccuracies.PredictX(float64(accuracy)))
	predictedRound := int32(predictedRoundFloat64)

	return predictedRound
}

func (pp *PerformancePrediction) PredictLoss(round int32) float32 {
	return float32(pp.regressionFunctionLosses.PredictY(float64(round)))
}

func (pp *PerformancePrediction) PredictRoundForLoss(loss float32) int32 {
	predictedRoundFloat64 := math.Ceil(pp.regressionFunctionLosses.PredictX(float64(loss)))
	predictedRound := int32(predictedRoundFloat64)

	return predictedRound
}

func (pp *PerformancePrediction) PrintPrediction() string {
	return pp.regressionFunctionAccuracies.PrintFunction()
}

func prepareXAndY(accuracies []float32) ([]float64, []float64) {
	xs := make([]float64, len(accuracies))
	ys := make([]float64, len(accuracies))

	for i, accuracy := range accuracies {
		xs[i] = float64(i)
		ys[i] = float64(accuracy)
	}

	return xs, ys
}
