package florch

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

func (orch *FlOrchestrator) calculateDatasetBasedScores() {
	var totalDatasetSize int64 = 0
	for _, client := range orch.configuration.Clients {
		totalDatasetSize += getDatasetSize(client.DataDistribution)
	}

	for _, client := range orch.configuration.Clients {
		clientDatasetSize := getDatasetSize(client.DataDistribution)
		overallDistribution := getOverallDataDistribution(orch.configuration.Clients, false, "")
		overallDistributionWithoutClient := getOverallDataDistribution(orch.configuration.Clients, true, client.Id)

		client.ClientUtility = model.ClientUtility{
			DatasetSizeScore:      float32(clientDatasetSize) / float32(totalDatasetSize),
			DataDistributionScore: float32(klDivergence(overallDistribution, overallDistributionWithoutClient)),
		}
	}
}

func (orch *FlOrchestrator) updateModelDifference() error {
	for _, client := range orch.configuration.Clients {
		logs, err := orch.contOrch.GetClientLogs(client.Id)
		if err != nil {
			return err
		}

		modelDifference := getModelDifferenceFromLogs(logs.String())
		client.ClientUtility.ModelDifference = append(client.ClientUtility.ModelDifference, modelDifference)
		client.ClientUtility.ModelDifferenceScore = float32(common.CalculateAverageFloat64(client.ClientUtility.ModelDifference))
	}

	return nil
}

// samples per class
func (orch *FlOrchestrator) getDataDistributionPerClient() error {
	for _, client := range orch.configuration.Clients {
		dataDistribution := make([]int, 10)
		logs, err := orch.contOrch.GetClientLogs(client.Id)
		if err != nil {
			return err
		}

		dataDistribution = getDataDistributionPerClientFromLogs(logs.String())
		client.ClientUtility.DataDistribution = dataDistribution
	}

	return nil
}

func getDatasetSize(dataDistribution map[string]int64) int64 {
	var datasetSize int64 = 0
	for _, samples := range dataDistribution {
		datasetSize += samples
	}
	return datasetSize
}

func getClientDistribution(client *model.FlClient) []float64 {
	totalSamples := 0
	samplesPerClass := make([]int64, 10)
	dataDistribution := client.DataDistribution
	for class, samples := range dataDistribution {
		i, _ := strconv.Atoi(class)
		samplesPerClass[i] += samples
		totalSamples += int(samples)
	}

	clientDistribution := make([]float64, 10)
	for i, samples := range samplesPerClass {
		percentage := float64(samples) / float64(totalSamples)
		if percentage == 0.0 {
			percentage = 0.0001
		}
		clientDistribution[i] = percentage
	}

	return clientDistribution
}

func getOverallDataDistribution(clients []*model.FlClient, skipClient bool, skipClientId string) []float64 {
	totalSamples := 0
	samplesPerClass := make([]int64, 10)
	for _, client := range clients {
		if skipClient && client.Id == skipClientId {
			continue
		}
		dataDistribution := client.DataDistribution
		for class, samples := range dataDistribution {
			i, _ := strconv.Atoi(class)
			samplesPerClass[i] += samples
			totalSamples += int(samples)
		}
	}

	overallDistribution := make([]float64, 10)
	for i, samples := range samplesPerClass {
		percentage := float64(samples) / float64(totalSamples)
		if percentage == 0.0 {
			percentage = 0.0001
		}
		overallDistribution[i] = percentage
	}

	return overallDistribution
}

func klDivergence(p, q []float64) float64 {
	if len(p) != len(q) {
		panic("Distributions must have the same number of parameters")
	}

	klDiv := 0.0
	for i := 0; i < len(p); i++ {
		if q[i] == 0 {
			continue
		}
		klDiv += p[i] * math.Log(p[i]/q[i])
	}
	return klDiv
}

func getModelDifferenceFromLogs(logs string) float64 {
	pattern := `Model difference \(L2-norm\) after training: ([\d.]+)`
	re := regexp.MustCompile(pattern)

	// Find the latest match in the buffer
	matches := re.FindAllStringSubmatch(logs, -1)

	if len(matches) > 0 {
		latestMatch := matches[len(matches)-1]
		value, _ := strconv.ParseFloat(latestMatch[1], 64)
		return value
	}

	return -1.0
}

func getDataDistributionPerClientFromLogs(logs string) []int {
	// Matches: Class counts (0..9): [numbers, possibly spaces]
	pattern := `Class counts...: \[([0-9,\s]+)\]`
	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringSubmatch(logs, -1)
	if len(matches) == 0 {
		return nil
	}

	// Take the latest match
	lastGroup := matches[len(matches)-1][1] // the inside of [...]
	parts := strings.Split(lastGroup, ",")
	out := make([]int, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			// skip any malformed piece
			continue
		}
		out = append(out, n)
	}
	return out
}
