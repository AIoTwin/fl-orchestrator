package florch

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/cost"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/flconfig"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/performance"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"github.com/hashicorp/go-hclog"
)

type FlOrchestrator struct {
	contOrch                 contorch.IContainerOrchestrator
	configurationModel       flconfig.IFlConfigurationModel
	eventBus                 *events.EventBus
	logger                   hclog.Logger
	resultsFileName          string
	nodesMap                 map[string]*model.Node
	batchSize                int32
	learningRate             float32
	configuration            *flconfig.FlConfiguration
	modelSize                float32
	costConfiguration        *cost.CostCofiguration
	progress                 *FlProgress
	reconfigurationEvaluator *ReconfigurationEvaluator
	rvaEnabled               bool
}

type FlProgress struct {
	globalRound          int32
	accuracies           []float32
	losses               []float32
	accuracyHasConverged bool
	communicationCost    float32
	costPerGlobalRound   float32
}

func NewFlOrchestrator(contOrch contorch.IContainerOrchestrator, eventBus *events.EventBus, logger hclog.Logger,
	configurationModelName string, epochs int32, localRounds int32, batchSize int32, learningRate float32,
	modelSize float32, costConfiguration *cost.CostCofiguration, rvaEnabled bool) (*FlOrchestrator, error) {
	orch := &FlOrchestrator{
		contOrch:                 contOrch,
		eventBus:                 eventBus,
		logger:                   logger,
		batchSize:                batchSize,
		learningRate:             learningRate,
		modelSize:                modelSize,
		costConfiguration:        costConfiguration,
		rvaEnabled:               rvaEnabled,
		reconfigurationEvaluator: &ReconfigurationEvaluator{isActive: false},
	}

	switch configurationModelName {
	case flconfig.MinimizeKld_ConfigModelName:
		orch.configurationModel = flconfig.NewMinimizeKldConfiguration(epochs, localRounds)
	case flconfig.MinimizeCommCost_ConfigModelName:
		orch.configurationModel = flconfig.NewMinimizeCommCostConfiguration(epochs, localRounds, modelSize)
	case flconfig.Cent_Hier_ConfigModelName:
		orch.configurationModel = flconfig.NewCentrHierFlConfiguration(modelSize, costConfiguration.CommunicationBudget)
	default:
		err := fmt.Errorf("invalid config model: %s", configurationModelName)
		return nil, err
	}

	return orch, nil
}

func (orch *FlOrchestrator) Start() error {
	nodesMap, err := orch.contOrch.GetAvailableNodes(true)
	if err != nil {
		orch.logger.Error(err.Error())
		return err
	}
	orch.nodesMap = nodesMap

	// set cofiguration and deploy FL
	orch.configuration = orch.configurationModel.GetOptimalConfiguration(nodesMapToArray(orch.nodesMap))
	orch.printConfiguration()
	orch.deployFl()

	orch.progress = &FlProgress{
		globalRound:       1,
		accuracies:        []float32{},
		losses:            []float32{},
		communicationCost: 0.0,
	}
	go orch.monitorFlProgress()

	nodeStateChangeChan := make(chan events.Event)
	orch.eventBus.Subscribe(common.NODE_STATE_CHANGE_EVENT_TYPE, nodeStateChangeChan)
	go orch.nodeStateChangeHandler(nodeStateChangeChan)

	go orch.contOrch.StartNodeStateChangeNotifier()

	flFinishedChan := make(chan events.Event)
	orch.eventBus.Subscribe(common.FL_FINISHED_EVENT_TYPE, flFinishedChan)
	go orch.flFinishedHandler(flFinishedChan)

	orch.resultsFileName = getResultsFileName()

	return nil
}

func (orch *FlOrchestrator) Stop() {
	orch.contOrch.StopAllNotifiers()
	orch.removeFl()
}

func (orch *FlOrchestrator) deployFl() {
	orch.deployGlobalAggregator(orch.configuration.GlobalAggregator)
	time.Sleep(20 * time.Second)

	for _, localAggregator := range orch.configuration.LocalAggregators {
		orch.deployLocalAggregator(localAggregator)
		time.Sleep(1 * time.Second)
	}
	time.Sleep(30 * time.Second)

	for _, client := range orch.configuration.Clients {
		client.BatchSize = orch.batchSize
		client.LearningRate = orch.learningRate
		orch.deployFlClient(client)
		time.Sleep(1 * time.Second)
	}
}

func (orch *FlOrchestrator) removeFl() {
	for _, client := range orch.configuration.Clients {
		orch.contOrch.RemoveClient(client)
	}

	for _, localAggregator := range orch.configuration.LocalAggregators {
		orch.contOrch.RemoveLocalAggregator(localAggregator)
	}

	orch.contOrch.RemoveGlobalAggregator(orch.configuration.GlobalAggregator)
}

func (orch *FlOrchestrator) reconfigure(newConfiguration *flconfig.FlConfiguration) {
	orch.logger.Info("Starting reconfiguration:")
	orch.printConfiguration()

	oldConfiguration := orch.configuration

	for _, oldClient := range oldConfiguration.Clients {
		newClient := common.GetClientInArray(newConfiguration.Clients, oldClient.Id)
		if (newClient == &model.FlClient{}) {
			orch.contOrch.RemoveClient(oldClient)
			orch.logger.Info(fmt.Sprintf("Removed client: %s", oldClient.Id))
		} else if oldClient.ParentNodeId != newClient.ParentNodeId {
			orch.contOrch.RemoveClient(oldClient)
			orch.logger.Info(fmt.Sprintf("Client changing cluster: %s", oldClient.Id))
		}
	}

	time.Sleep(5 * time.Second)
	orch.logger.Info("Deploying new configuration...")

	for _, newClient := range newConfiguration.Clients {
		oldClient := common.GetClientInArray(oldConfiguration.Clients, newClient.Id)
		if (oldClient == &model.FlClient{}) {
			orch.deployFlClient(newClient)
		} else if oldClient.ParentAddress != newClient.ParentAddress {
			orch.deployFlClient(newClient)
		}
	}

	orch.configuration = newConfiguration
}

func (orch *FlOrchestrator) deployGlobalAggregator(flAggregator *model.FlAggregator) error {
	globalAggregatorConfigFilesData, err := BuildGlobalAggregatorConfigFiles(flAggregator)
	if err != nil {
		orch.logger.Error(fmt.Sprintf("Error while initializing global aggregator config files: %s", err.Error()))
		return err
	}

	if err := orch.contOrch.CreateGlobalAggregator(flAggregator, globalAggregatorConfigFilesData); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while deploying global aggregator: %s", err.Error()))
		return err
	}

	orch.logger.Info("Global aggregator deployed!")

	return nil
}

func (orch *FlOrchestrator) deployLocalAggregator(flAggregator *model.FlAggregator) error {
	localAggregatorConfigFilesData, err := BuildLocalAggregatorConfigFiles(flAggregator)
	if err != nil {
		orch.logger.Error(fmt.Sprintf("Error while initializing local aggregator config files: %s", err.Error()))
		return err
	}

	if err := orch.contOrch.CreateLocalAggregator(flAggregator, localAggregatorConfigFilesData); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while deploying local aggregator: %s", err.Error()))
		return err
	}

	orch.logger.Info("Local aggregator deployed!")

	return nil
}

func (orch *FlOrchestrator) deployFlClient(client *model.FlClient) error {
	clientConfigFilesData, err := BuildClientConfigFiles(client)
	if err != nil {
		orch.logger.Error(fmt.Sprintf("Error while initializing client %s config files: %s", client.Id, err.Error()))
		return err
	}

	if err := orch.contOrch.CreateFlClient(client, clientConfigFilesData); err != nil {
		orch.logger.Error(fmt.Sprintf("Error while creating client %s deployment: %s", client.Id, err.Error()))
		return err
	}

	orch.logger.Info(fmt.Sprintf("Client %s deployed!", client.Id))

	return nil
}

func (orch *FlOrchestrator) flFinishedHandler(eventChan <-chan events.Event) {
	for event := range eventChan {
		flFinishedEvent, ok := event.Data.(events.FlFinishedEvent)
		if !ok {
			orch.logger.Info("Invalid event data")
			continue
		}

		orch.logger.Info(fmt.Sprintf("FL finished! Exit message: %s", flFinishedEvent.ExitMessage))
		orch.removeFl()

		return
	}
}

func (orch *FlOrchestrator) nodeStateChangeHandler(eventChan <-chan events.Event) {
	for event := range eventChan {
		nodeStateChangeEvent, ok := event.Data.(events.NodeStateChangeEvent)
		if !ok {
			orch.logger.Info("Invalid event data")
			continue
		}

		// Handle the event
		orch.logger.Info("New event:")
		orch.logger.Info(fmt.Sprintf("Nodes added: %v", nodeStateChangeEvent.NodesAdded))
		orch.logger.Info(fmt.Sprintf("Node removed: %v", nodeStateChangeEvent.NodesRemoved))

		for _, node := range nodeStateChangeEvent.NodesAdded {
			orch.nodesMap[node.Id] = node
		}

		for _, node := range nodeStateChangeEvent.NodesRemoved {
			delete(orch.nodesMap, node.Id)
		}

		orch.runReconfigurationModel()
	}
}

func (orch *FlOrchestrator) runReconfigurationModel() {
	finishedGlobalRound := orch.progress.globalRound - 2

	newConfiguration := orch.configurationModel.GetOptimalConfiguration(nodesMapToArray(orch.nodesMap))
	newConfigCost := cost.GetGlobalRoundCost(newConfiguration, orch.nodesMap, orch.modelSize)

	reconfigurationChangeCost := cost.GetReconfigurationChangeCost(orch.configuration, newConfiguration, orch.nodesMap, orch.modelSize)
	orch.logger.Info(fmt.Sprintf("Reconfiguration change cost: %.2f", reconfigurationChangeCost))

	postReconfigurationCost := newConfigCost - orch.progress.costPerGlobalRound
	orch.logger.Info(fmt.Sprintf("Post reconfiguration cost: %.2f", postReconfigurationCost))

	if orch.rvaEnabled {
		if orch.costConfiguration.CostType == cost.TotalBudget_CostType {
			budgetRemaning := orch.costConfiguration.CommunicationBudget - orch.progress.communicationCost
			roundsRemainingCurrent := math.Floor(float64(budgetRemaning / orch.progress.costPerGlobalRound))
			roundsRemainingNew := math.Floor(float64((budgetRemaning - reconfigurationChangeCost) / newConfigCost))
			if roundsRemainingNew < roundsRemainingCurrent {
				ppCurrent := performance.NewPerformancePrediction(orch.progress.accuracies, orch.progress.losses,
					performance.LogarithmicRegression_PredictionType, 0)
				orch.reconfigurationEvaluator = &ReconfigurationEvaluator{
					isActive:          true,
					evaluationRound:   finishedGlobalRound + ReconfEvalWindow,
					startAccuracy:     orch.progress.accuracies[len(orch.progress.accuracies)-1],
					startLoss:         orch.progress.losses[len(orch.progress.losses)-1],
					startConfig:       orch.configuration,
					startPp:           ppCurrent,
					startCostPerRound: orch.progress.costPerGlobalRound,
					endConfig:         newConfiguration,
					endPp:             ppCurrent,
					endAccuracies:     []float32{},
					endLosses:         []float32{},
				}
				orch.logger.Info(fmt.Sprintf("reconfiguration evaluation set for round: %d", orch.reconfigurationEvaluator.evaluationRound))
			}
		} else if orch.costConfiguration.CostType == cost.CostMinimization_CostType {
			ppCurrent := performance.NewPerformancePrediction(orch.progress.accuracies, orch.progress.losses,
				performance.LogarithmicRegression_PredictionType, 0)
			roundPredicted := ppCurrent.PredictRoundForAccuracy(orch.costConfiguration.TargetAccuracy)
			roundsRemainingCurrent := float32(roundPredicted - finishedGlobalRound)
			costRemainingCurrent := float32(roundsRemainingCurrent) * orch.progress.costPerGlobalRound
			roundsRemainingNew := (costRemainingCurrent - reconfigurationChangeCost) / newConfigCost
			if roundsRemainingNew < roundsRemainingCurrent {
				orch.reconfigurationEvaluator = &ReconfigurationEvaluator{
					isActive:          true,
					evaluationRound:   finishedGlobalRound + ReconfEvalWindow,
					startAccuracy:     orch.progress.accuracies[len(orch.progress.accuracies)-1],
					startLoss:         orch.progress.losses[len(orch.progress.losses)-1],
					startConfig:       orch.configuration,
					startPp:           ppCurrent,
					startCostPerRound: orch.progress.costPerGlobalRound,
					endConfig:         newConfiguration,
					endPp:             ppCurrent,
					endAccuracies:     []float32{},
					endLosses:         []float32{},
				}
				orch.logger.Info(fmt.Sprintf("reconfiguration evaluation set for round: %d", orch.reconfigurationEvaluator.evaluationRound))
			}
		}
	}

	orch.progress.communicationCost += reconfigurationChangeCost
	orch.reconfigure(newConfiguration)
}

func (orch *FlOrchestrator) monitorFlProgress() {
	for {
		logsBuffer, err := orch.contOrch.GetGlobalAggregatorLogs()
		if err != nil {
			orch.logger.Error(fmt.Sprintf("Error while obtaining GA logs: %s", err.Error()))
			time.Sleep(1 * time.Second)
			continue
		}

		logs := logsBuffer.String()
		if strings.Contains(logs, fmt.Sprintf("fit_round %d:", orch.progress.globalRound)) {
			finishedGlobalRound := orch.progress.globalRound - 1

			orch.logger.Info(fmt.Sprintf("Finished global round %d", finishedGlobalRound))

			accuracy := getLatestAccuracyFromLogs(logs)
			orch.progress.accuracies = append(orch.progress.accuracies, accuracy)
			orch.logger.Info(fmt.Sprintf("Latest accuracy: %.2f", accuracy))

			loss := getLatestLossFromLogs(logs, finishedGlobalRound)
			orch.progress.losses = append(orch.progress.losses, loss)
			orch.logger.Info(fmt.Sprintf("Latest loss: %.2f", loss))

			orch.progress.costPerGlobalRound = cost.GetGlobalRoundCost(orch.configuration, orch.nodesMap, orch.modelSize)

			if finishedGlobalRound > 0 {
				orch.logger.Info(fmt.Sprintf("Cost per global round: %.2f", orch.progress.costPerGlobalRound))
				orch.progress.communicationCost += orch.progress.costPerGlobalRound
				orch.logger.Info(fmt.Sprintf("Current total cost: %.2f", orch.progress.communicationCost))

				orch.progress.accuracyHasConverged = hasConverged(orch.progress.accuracies, 0.1, 5, 3)
				if orch.progress.accuracyHasConverged {
					orch.logger.Info("Accuracy has converged!")
				}
			}

			orch.logger.Info(fmt.Sprintf("Started global round %d", orch.progress.globalRound))

			writeResultsToFile(orch.resultsFileName, finishedGlobalRound, accuracy, loss, orch.progress.communicationCost)

			if orch.costConfiguration.CostType == cost.TotalBudget_CostType {
				if orch.progress.communicationCost >= orch.costConfiguration.CommunicationBudget {
					orch.logger.Info(fmt.Sprintf("Communication budget exceeded!\nTotal cost: %.2f\nFinal accuracy: %.2f",
						orch.progress.communicationCost, accuracy))
					orch.removeFl()
					break
				}
			} else if orch.costConfiguration.CostType == cost.CostMinimization_CostType {
				if accuracy >= orch.costConfiguration.TargetAccuracy {
					orch.logger.Info(fmt.Sprintf("Target accuracy reached!\nTotal cost: %.2f\nFinal accuracy: %.2f",
						orch.progress.communicationCost, accuracy))
					orch.removeFl()
					break
				}
			}

			if orch.reconfigurationEvaluator.isActive {
				orch.reconfigurationEvaluator.endAccuracies = append(orch.reconfigurationEvaluator.endAccuracies, accuracy)
				orch.reconfigurationEvaluator.endLosses = append(orch.reconfigurationEvaluator.endLosses, loss)

				if finishedGlobalRound == orch.reconfigurationEvaluator.evaluationRound {
					orch.evaluateReconfiguration()
				}
			}

			if finishedGlobalRound == 10 {
				orch.logger.Info("Applying changes...")
				applyChanges("../../configs/cluster/cluster.csv", "../../configs/cluster/changes.csv")
			}

			orch.progress.globalRound++
		}

		time.Sleep(10 * time.Second)
	}
}

func (orch *FlOrchestrator) printConfiguration() {
	configToPrint := ""

	configToPrint += fmt.Sprintln("Global aggregator ::")
	globalAggregator := orch.configuration.GlobalAggregator
	configToPrint += fmt.Sprintf("\tNode id:%s\t| Rounds:%d\n", globalAggregator.Id, globalAggregator.Rounds)
	configToPrint += fmt.Sprintln("Local aggregators ::")
	for _, a := range orch.configuration.LocalAggregators {
		configToPrint += fmt.Sprintf("\tNode id:%s\t| Parent address:%s\t| Local rounds:%d Rounds:%d\n", a.Id, a.ParentAddress,
			a.LocalRounds, a.Rounds)
	}
	configToPrint += fmt.Sprintln("Clients ::")
	for _, c := range orch.configuration.Clients {
		configToPrint += fmt.Sprintf("\tNode id:%s\t| Parent node:%s\t| Epochs:%d\n", c.Id, c.ParentNodeId, c.Epochs)
	}
	configToPrint += fmt.Sprintln("Epochs: ", orch.configuration.Epochs)
	configToPrint += fmt.Sprintln("Local rounds: ", orch.configuration.LocalRounds)

	orch.logger.Info(configToPrint)
}

// HELPERS

func getLatestAccuracyFromLogs(logs string) float32 {
	pattern := `accuracy': ([0-9]*\.[0-9]+)`
	r := regexp.MustCompile(pattern)

	matches := r.FindAllStringSubmatch(logs, -1)

	if len(matches) > 0 {
		latestMatch := matches[len(matches)-1]
		accuracy, _ := strconv.ParseFloat(latestMatch[1], 32)
		return float32(accuracy)
	}

	return -1.0
}

func getLatestLossFromLogs(logs string, finishedGlobalRound int32) float32 {
	// Define the regex patterns
	patterns := []string{
		`\(loss, other metrics\): ([\d.]+),`,
		`fit progress: \(\d+, ([\d.]+),`,
	}

	var re *regexp.Regexp
	if finishedGlobalRound == 0 {
		re = regexp.MustCompile(patterns[0])
	} else {
		re = regexp.MustCompile(patterns[1])
	}

	// Find all matches in the string
	matches := re.FindAllStringSubmatch(logs, -1)

	if len(matches) > 0 {
		latestMatch := matches[len(matches)-1]
		loss, _ := strconv.ParseFloat(latestMatch[1], 32)
		return float32(loss)
	}

	return -1.0
}

func movingAverage(values []float32, windowSize int) []float32 {
	if len(values) < windowSize {
		return nil // Not enough data for the window size
	}
	averages := make([]float32, len(values)-windowSize+1)
	for i := 0; i <= len(values)-windowSize; i++ {
		sum := float32(0.0)
		for j := i; j < i+windowSize; j++ {
			sum += values[j]
		}
		averages[i] = sum / float32(windowSize)
	}
	return averages
}

func hasConverged(accuracies []float32, threshold float32, patience int, windowSize int) bool {
	averages := movingAverage(accuracies, windowSize)
	if len(averages) < patience+1 {
		return false // Not enough data to determine convergence
	}

	for i := len(averages) - patience; i < len(averages); i++ {
		improvement := averages[i] - averages[i-1]
		if math.Abs(float64(improvement)) > float64(threshold) {
			return false // If improvement is greater than the threshold, no convergence
		}
	}
	return true // Converged if all improvements are below the threshold
}

func nodesMapToArray(nodesMap map[string]*model.Node) []*model.Node {
	nodesArray := make([]*model.Node, 0, len(nodesMap))

	for _, node := range nodesMap {
		nodesArray = append(nodesArray, node)
	}

	return nodesArray
}

func getResultsFileName() string {
	return fmt.Sprintf("../../experiments/results/results_%s.csv", time.Now().Format("2006-01-02_15-04"))
}

func writeResultsToFile(fileName string, round int32, accuracy float32, loss float32, cost float32) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Convert the float values to strings and store them in the CSV
	record := []string{fmt.Sprintf("%d", round), fmt.Sprintf("%.2f", accuracy), fmt.Sprintf("%.2f", loss),
		fmt.Sprintf("%.2f", cost)}
	if err := writer.Write(record); err != nil {
		fmt.Printf("Failed to write record: %v\n", err)
		return
	}
}

func applyChanges(clusterFileName string, changesFileName string) {
	source, err := os.Open(changesFileName)
	if err != nil {
		fmt.Printf("Error opening source file: %v\n", err)
		return
	}
	defer source.Close()

	dest, err := os.OpenFile(clusterFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error opening destination file: %v\n", err)
		return
	}
	defer dest.Close()

	if _, err = dest.WriteString("\n"); err != nil {
		fmt.Printf("Error writing new line: %v\n", err)
		return
	}

	reader := bufio.NewReader(source)
	if _, err = io.Copy(dest, reader); err != nil {
		fmt.Printf("Error appending content: %v\n", err)
		return
	}

	fmt.Println("Content appended successfully.")
}
