package florch

import (
	"fmt"
	"math"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/cost"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/flconfig"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/performance"
)

const ReconfEvalWindow = 5

type ReconfigurationEvaluator struct {
	isActive          bool
	evaluationRound   int32
	startAccuracy     float32
	startLoss         float32
	startConfig       *flconfig.FlConfiguration
	startPp           *performance.PerformancePrediction
	startCostPerRound float32
	endConfig         *flconfig.FlConfiguration
	endPp             *performance.PerformancePrediction
	endAccuracies     []float32
	endLosses         []float32
}

func (orch *FlOrchestrator) evaluateReconfiguration() {
	orch.logger.Info("Starting reconf evaluation...")

	finishedGlobalRound := orch.progress.globalRound - 1

	endAccuracy := orch.progress.accuracies[len(orch.progress.accuracies)-1]
	orch.logger.Info(fmt.Sprintf("Dest accuracy: %.2f", endAccuracy))

	predictedAccuracy := orch.reconfigurationEvaluator.startPp.PredictAccuracy(finishedGlobalRound)
	orch.logger.Info(fmt.Sprintf("Src predicted accuracy: %.2f", predictedAccuracy))

	reconfigurationChangeCost := cost.GetReconfigurationChangeCost(orch.reconfigurationEvaluator.endConfig, orch.reconfigurationEvaluator.startConfig,
		orch.nodesMap, orch.modelSize)
	orch.logger.Info(fmt.Sprintf("Reconf change cost: %.2f", reconfigurationChangeCost))

	orch.reconfigurationEvaluator.endPp = performance.NewPerformancePrediction(orch.reconfigurationEvaluator.endAccuracies,
		orch.reconfigurationEvaluator.endLosses, performance.LogarithmicRegression_PredictionType,
		int(orch.reconfigurationEvaluator.evaluationRound)-ReconfEvalWindow)

	if orch.costConfiguration.CostType == cost.TotalBudget_CostType {
		budgetRemaning := orch.costConfiguration.CommunicationBudget - orch.progress.communicationCost
		orch.logger.Info(fmt.Sprintf("Remaining budget: %.2f", budgetRemaning))

		startRoundsRemaining := math.Floor(float64((budgetRemaning - reconfigurationChangeCost) / orch.reconfigurationEvaluator.startCostPerRound))
		orch.logger.Info(fmt.Sprintf("Src rounds remaining: %.2f", startRoundsRemaining))

		endRoundsRemaining := math.Floor(float64(budgetRemaning / orch.progress.costPerGlobalRound))
		orch.logger.Info(fmt.Sprintf("Dest rounds remaining: %.2f", endRoundsRemaining))

		orch.logger.Info(fmt.Sprintf("Src predicted function: %s", orch.reconfigurationEvaluator.startPp.PrintPrediction()))

		startAccuracyFinal := orch.reconfigurationEvaluator.startPp.PredictAccuracy(finishedGlobalRound + int32(startRoundsRemaining))
		orch.logger.Info(fmt.Sprintf("Src accuracy final: %.2f", startAccuracyFinal))

		orch.logger.Info(fmt.Sprintf("Dest predicted function: %s", orch.reconfigurationEvaluator.endPp.PrintPrediction()))

		endAccuracyFinal := orch.reconfigurationEvaluator.endPp.PredictAccuracy(finishedGlobalRound + int32(endRoundsRemaining))
		orch.logger.Info(fmt.Sprintf("Dest accuracy final: %.2f", endAccuracyFinal))

		if startAccuracyFinal > endAccuracyFinal {
			orch.logger.Info("Reconfiguration introduces performance degradation. Reverting configuration...")
			go orch.reconfigure(orch.reconfigurationEvaluator.startConfig)
			orch.reconfigurationEvaluator = &ReconfigurationEvaluator{isActive: false}
			return
		}
	} else if orch.costConfiguration.CostType == cost.CostMinimization_CostType {
		targetAccuracy := orch.costConfiguration.TargetAccuracy

		startRoundsRemaining := orch.reconfigurationEvaluator.startPp.PredictRoundForAccuracy(targetAccuracy) -
			orch.progress.globalRound
		orch.logger.Info(fmt.Sprintf("Src rounds remaining: %d", startRoundsRemaining))

		endRoundsRemaining := orch.reconfigurationEvaluator.endPp.PredictRoundForAccuracy(targetAccuracy) -
			orch.progress.globalRound
		orch.logger.Info(fmt.Sprintf("Dest rounds remaining: %.d", endRoundsRemaining))

		startCostRemaining := reconfigurationChangeCost + float32(startRoundsRemaining)*orch.reconfigurationEvaluator.startCostPerRound
		orch.logger.Info(fmt.Sprintf("Src cost remaining: %.2f", startCostRemaining))

		endCostRemaining := float32(endRoundsRemaining) * orch.progress.costPerGlobalRound
		orch.logger.Info(fmt.Sprintf("End cost remaining: %.2f", endCostRemaining))

		if startCostRemaining < endCostRemaining {
			orch.logger.Info("Reconfiguration introduces performance degradation. Reverting configuration...")
			go orch.reconfigure(orch.reconfigurationEvaluator.startConfig)
			orch.reconfigurationEvaluator = &ReconfigurationEvaluator{isActive: false}
			return
		}
	}

	orch.logger.Info("Reconfiguration introduces performance improvement. Continuing with new configuration...")
}
