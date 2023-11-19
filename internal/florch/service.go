package florch

import (
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildGlobalAggregatorService(globalAggregator model.GlobalAggregator) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GLOBAL_AGGREGATOR_SERVICE_NAME,
			Namespace: corev1.NamespaceDefault,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"fl": "ga",
			},
			Ports: []corev1.ServicePort{
				{
					Port: globalAggregator.Port,
				},
			},
		},
	}

	return service
}
