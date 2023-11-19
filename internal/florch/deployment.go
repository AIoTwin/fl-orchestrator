package florch

import (
	"fmt"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildGlobalAggregatorDeployment(globalAggregator model.GlobalAggregator) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: GLOBAL_AGGRETATOR_DEPLOYMENT_NAME,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"fl": "ga",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"fl": "ga",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "fl-ga",
							Image: GLOBAL_AGGRETATOR_IMAGE,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: globalAggregator.Port,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: GLOBAL_AGGRETATOR_MOUNT_PATH,
									Name:      "gaconfig",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "gaconfig",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									Items: []corev1.KeyToPath{},
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "gacm",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment
}

func BuildClientDeployment(client model.FlClient) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: getClientDeploymentName(client),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"fl": fmt.Sprintf("client-%s", client.Id),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"fl": fmt.Sprintf("client-%s", client.Id),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "fl-client",
							Image: FL_CLIENT_IMAGE,
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: FL_CLIENT_CONFIG_MOUNT_PATH,
									Name:      "clientconfig",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "clientconfig",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									Items: []corev1.KeyToPath{},
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("clientcm-%s", client.Id),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment
}

func getClientDeploymentName(client model.FlClient) string {
	return fmt.Sprintf("%s-%s", FL_CLIENT_DEPLOYMENT_PREFIX, client.Id)
}
