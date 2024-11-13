package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"gonum.org/v1/gonum/mat"
)

// LogarithmicRegression performs a logarithmic fit
func LogarithmicRegression(xs, ys []float64) (float64, float64) {
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
		return 0, 0
	}

	// Coefficients for the logarithmic function
	a := coef.AtVec(0)
	b := coef.AtVec(1)

	return a, b
}

// PredictLogarithmic predicts a value for x using the logarithmic model
func PredictLogarithmic(a, b, x float64) float64 {
	return a + b*math.Log(x+1)
}

// PredictXForY solves for x given a y value using the logarithmic model
func PredictXForY(a, b, y float64) float64 {
	if b == 0 {
		return math.NaN() // Avoid division by zero
	}
	return math.Exp((y-a)/b) - 1
}

func main() {
	records := common.ReadCsvFile("./test.csv")

	// Initialize slices for the first two columns
	xs := make([]float64, len(records))
	ys := make([]float64, len(records))

	// Process each line
	for i, record := range records {
		xs[i], _ = strconv.ParseFloat(record[0], 64)
		ys[i], _ = strconv.ParseFloat(record[1], 64)
	}

	// Perform logarithmic regression
	a, b := LogarithmicRegression(xs, ys)

	// Display the coefficients
	fmt.Printf("Logarithmic coefficients: a = %.3f, b = %.3f\n", a, b)

	// Predict a value for x = 20
	predictX := 57.0
	predictY := PredictLogarithmic(a, b, predictX)

	fmt.Printf("Predicted value at x=%.2f: %.2f\n", predictX, predictY)

	// Predict a value for x = 20
	predictY = 100.0
	predictX = PredictXForY(a, b, predictY)

	fmt.Printf("Predicted value at y=%.2f: %.2f\n", predictY, predictX)
}
