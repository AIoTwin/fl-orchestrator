package k8sorch

import (
	"fmt"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/common"
	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildGlobalAggregatorDeployment(aggregator *model.FlAggregator) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.GLOBAL_AGGRETATOR_DEPLOYMENT_NAME,
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
							Image: common.GLOBAL_AGGRETATOR_IMAGE,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: aggregator.Port,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "gaconfig",
									MountPath: "/home/task.py",
									SubPath:   "task.py",
								},
								{
									Name:      "gaconfig",
									MountPath: "/home/global_server_config.yaml",
									SubPath:   "global_server_config.yaml",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1.0"),
									corev1.ResourceMemory: resource.MustParse("1500Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2.0"),
									corev1.ResourceMemory: resource.MustParse("2000Mi"),
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
										Name: common.GetGlobalAggregatorConfigMapName(aggregator.Id),
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

func BuildLocalAggregatorDeployment(aggregator *model.FlAggregator) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.GetLocalAggregatorDeploymentName(aggregator.Id),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"fl": fmt.Sprintf("la-%s", aggregator.Id),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"fl": fmt.Sprintf("la-%s", aggregator.Id),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "fl-la",
							Image: common.LOCAL_AGGRETATOR_IMAGE,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: aggregator.Port,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "laconfig",
									MountPath: "/home/local_server_config.yaml",
									SubPath:   "local_server_config.yaml",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1.0"),
									corev1.ResourceMemory: resource.MustParse("1500Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2.0"),
									corev1.ResourceMemory: resource.MustParse("2000Mi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "laconfig",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									Items: []corev1.KeyToPath{},
									LocalObjectReference: corev1.LocalObjectReference{
										Name: common.GetLocalAggregatorConfigMapName(aggregator.Id),
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

func BuildClientDeployment(client *model.FlClient) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.GetClientDeploymentName(client.Id),
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
							Image: common.FL_CLIENT_IMAGE,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "clientconfig",
									MountPath: "/home/task.py",
									SubPath:   "task.py",
								},
								{
									Name:      "clientconfig",
									MountPath: "/home/client_config.yaml",
									SubPath:   "client_config.yaml",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1.0"),
									corev1.ResourceMemory: resource.MustParse("1500Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2.0"),
									corev1.ResourceMemory: resource.MustParse("2000Mi"),
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
										Name: common.GetClientConfigMapName(client.Id),
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
