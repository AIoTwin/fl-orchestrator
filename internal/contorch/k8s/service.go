package k8sorch

import (
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildGlobalAggregatorService(flAggregator *model.FlAggregator) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.GetAggregatorServiceName(flAggregator.Id),
			Namespace: corev1.NamespaceDefault,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"fl": "ga",
			},
			Ports: []corev1.ServicePort{
				{
					Port: flAggregator.Port,
				},
			},
		},
	}

	return service
}
