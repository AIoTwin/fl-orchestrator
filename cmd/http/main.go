package main

import (
	"io"
	"os"

	k8sorch "github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/contorch/k8s"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/events"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/server"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

func main() {
	_ = os.Mkdir("log", 0777)
	logFile, err := os.OpenFile("log/run.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			panic(err)
		}
	}()

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "fl-orch",
		Level:  hclog.LevelFromString("DEBUG"),
		Output: io.MultiWriter(os.Stdout, logFile),
	})

	eventBus := events.NewEventBus()

	simulation := false
	if len(os.Args) == 2 {
		deplType := os.Args[1]
		if deplType == "sim" {
			simulation = true
		}
	}

	k8sConfigFilePath := "../../configs/cluster/kube_config.yaml"
	k8sOrchestrator, err := k8sorch.NewK8sOrchestrator(k8sConfigFilePath, eventBus, simulation)
	if err != nil {
		logger.Error("Error while initializing k8s client ::", err.Error())
		return
	}

	handler := server.NewHandler(logger, eventBus, k8sOrchestrator)

	defaultRouter := mux.NewRouter()
	defaultRouter.HandleFunc("/fl/start", handler.StartFl)
	defaultRouter.HandleFunc("/fl/stop/{runId}", handler.StopFl)

	server.StartHttpServer(logger, defaultRouter)
}
