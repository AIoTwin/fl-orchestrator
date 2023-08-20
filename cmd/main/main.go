package main

import (
	"log"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/pkg/k8s"
)

func main() {
	k8sClient, err := k8s.NewK8sClient("/home/ivan/.kube/config")
	if err != nil {
		log.Fatal("Error while initializing k8s client ::", err.Error())
		return
	}

	k8sClient.GetNodesStatus()
}
