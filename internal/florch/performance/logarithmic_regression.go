package performance

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

type LogarithmicRegression struct {
	xs []float64
	ys []float64
	a  float64
	b  float64
}

// LogarithmicRegression performs a logarithmic fit
func NewLogarithmicRegression(xs, ys []float64) *LogarithmicRegression {
	lr := &LogarithmicRegression{
		xs: xs,
		ys: ys,
	}

	// Transform xs to logarithms
	logXs := make([]float64, len(xs))
	for i, x := range xs {
		logXs[i] = math.Log(x + 1) // log(x) with +1 to avoid log(0)
	}

	// Linear regression on transformed data
	const degree = 1
	X := mat.NewDense(len(xs), degree+1, nil)
	for i, x := range logXs {
		X.Set(i, 0, 1) // constant term
		X.Set(i, 1, x)
	}
	Y := mat.NewVecDense(len(ys), ys)

	var coef mat.VecDense
	err := coef.SolveVec(mat.Matrix(X), mat.Vector(Y))
	if err != nil {
		fmt.Println("Error solving the linear system:", err)
		return &LogarithmicRegression{}
	}

	// Coefficients for the logarithmic function
	a := coef.AtVec(0)
	b := coef.AtVec(1)

	lr.a = a
	lr.b = b

	return lr
}

// PredictY predicts a value for x using the logarithmic model
func (lr *LogarithmicRegression) PredictY(x float64) float64 {
	return lr.a + lr.b*math.Log(x+1)
}

// PredictX solves for x given a y value using the logarithmic model
func (lr *LogarithmicRegression) PredictX(y float64) float64 {
	if lr.b == 0 {
		return math.NaN() // Avoid division by zero
	}
	return math.Exp((y-lr.a)/lr.b) - 1
}

func (lr *LogarithmicRegression) PrintFunction() string {
	return fmt.Sprintf("f(x) = %.2f + %.2f * ln(x+1)", lr.a, lr.b)
}
