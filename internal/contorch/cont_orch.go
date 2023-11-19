package contorch

import (
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type IContainerOrchestrator interface {
	GetNodesStatus() ([]*model.Node, error)
	CreateConfigMapFromFiles(configMapName string, filesData map[string]string) error
	CreateDeployment(deployment *appsv1.Deployment) error
	DeleteDeployment(deploymentName string) error
	CreateService(service *corev1.Service) error
	DeleteService(serviceName string) error
}
