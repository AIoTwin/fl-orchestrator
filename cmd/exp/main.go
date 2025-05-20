package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	numClasses := 10
	numClients := 10
	samplesPerClass := 1000
	alpha := 0.5 // Adjust this for more/less non-IID

	// Seed random for consistent results
	rand.Seed(time.Now().UnixNano())

	// Generate distributions for each client
	clientDistributions := make([][]int, numClients)
	for i := 0; i < numClients; i++ {
		clientDistributions[i] = generateDirichletDistribution(numClasses, samplesPerClass, alpha)
	}

	// Write distributions to CSV file
	err := writeInitialClusterCSV("../../configs/cluster/cluster.csv", clientDistributions[:8])
	if err != nil {
		fmt.Println("Error writing CSV:", err)
		return
	}

	err = writeChangesCSV("../../configs/cluster/changes.csv", clientDistributions[8:])
	if err != nil {
		fmt.Println("Error writing CSV:", err)
		return
	}

	fmt.Println("Client distributions written to CSVs")
}

// Generates a Dirichlet distribution for a client with `numClasses` and `samplesPerClass`
func generateDirichletDistribution(numClasses int, samplesPerClass int, alpha float64) []int {
	weights := make([]float64, numClasses)
	sum := 0.0

	// Generate weights using gamma distribution for Dirichlet
	for i := 0; i < numClasses; i++ {
		weights[i] = rand.ExpFloat64() / alpha // Inversely scale with alpha
		sum += weights[i]
	}

	// Convert weights to sample counts for each class
	classSamples := make([]int, numClasses)
	currentTotal := 0
	for i := 0; i < numClasses; i++ {
		classSamples[i] = int((weights[i] / sum) * float64(samplesPerClass))
		currentTotal += classSamples[i]
	}

	// Adjust to match totalSamples exactly
	classSamples[0] += samplesPerClass - currentTotal

	return classSamples
}

func writeInitialClusterCSV(filename string, clientDistributions [][]int) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i, dist := range clientDistributions {
		clientId := i + 4
		clientData := fmt.Sprintf("n%d", clientId)

		flType := "client"

		commCosts := ""
		if i < 4 {
			commCosts = "n1:120,n2:20,n3:60"
		} else {
			commCosts = "n1:120,n2:60,n3:20"
		}

		classCounts := ""
		for classID, count := range dist {
			classCounts += fmt.Sprintf("%d:%d,", classID, count)
		}
		classCounts = classCounts[:len(classCounts)-1] // Remove trailing comma

		row := []string{clientData, flType, commCosts, classCounts}
		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeChangesCSV(filename string, clientDistributions [][]int) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i, dist := range clientDistributions {
		clientId := i + 12
		clientData := fmt.Sprintf("n%d", clientId)

		flType := "client"

		commCosts := ""
		if i == 0 {
			commCosts = "n1:120,n2:20,n3:60"
		} else {
			commCosts = "n1:120,n2:60,n3:20"
		}

		classCounts := ""
		for classID, count := range dist {
			classCounts += fmt.Sprintf("%d:%d,", classID, count)
		}
		classCounts = classCounts[:len(classCounts)-1] // Remove trailing comma

		row := []string{clientData, flType, commCosts, classCounts}
		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
