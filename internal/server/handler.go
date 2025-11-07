package server

import (
	"fmt"
	"net/http"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

type Handler struct {
	logger        hclog.Logger
	eventBus      *events.EventBus
	contOrch      contorch.IContainerOrchestrator
	orchestrators map[string]*florch.FlOrchestrator
}

func NewHandler(logger hclog.Logger, eventBus *events.EventBus, contOrch contorch.IContainerOrchestrator) *Handler {
	return &Handler{
		logger:        logger,
		eventBus:      eventBus,
		contOrch:      contOrch,
		orchestrators: map[string]*florch.FlOrchestrator{},
	}
}

func (handler *Handler) StartFl(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	runId := uuid.New().String()

	request := &StartFlRequest{}
	err := fromJSON(request, r.Body)
	if err != nil {
		handler.logger.Error("erorr starting FL: ", "error", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	flOrchestrator, err := florch.NewFlOrchestrator(handler.contOrch, handler.eventBus, handler.logger, request.ConfigurationModel,
		request.Epochs, request.LocalRounds, request.TrainingParams.BatchSize, request.TrainingParams.LearningRate,
		request.ModelSize, request.CostSource, &request.CostConfiguration, request.RvaEnabled)
	if err != nil {
		handler.logger.Error("erorr starting FL", "error", err)
		rw.WriteHeader(http.StatusBadRequest)
		toJSON("invalid configuration model", rw)
		return
	}

	handler.orchestrators[runId] = flOrchestrator

	handler.logger.Info(fmt.Sprintf("Starting FL with config %s, modelSize %f, and cost type %s", request.ConfigurationModel,
		request.ModelSize, request.CostConfiguration.CostType))

	err = flOrchestrator.Start()
	if err != nil {
		handler.logger.Error("erorr starting FL", "error", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	toJSON(runId, rw)
}

func (handler *Handler) StopFl(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	runId := getURLParameter(r, "runId")

	handler.logger.Info(fmt.Sprintf("Stopping FL with run ID: %s", runId))

	flOrchestrator := handler.orchestrators[runId]
	if flOrchestrator != nil {
		flOrchestrator.Stop()
		rw.WriteHeader(http.StatusOK)
	} else {
		rw.WriteHeader(http.StatusBadRequest)
		toJSON("no run with the given ID", rw)
	}
}

func getURLParameter(r *http.Request, parameter string) string {
	vars := mux.Vars(r)
	id := vars[parameter]
	return id
}
