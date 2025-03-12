package k8s

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"cmp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CreateOrUpdateDeploymentParams struct {
	Name                 string
	Namespace            string
	Entrypoint           []string
	ContainerPort        int32
	Image                string
	CopyFolderTo         string
	Replicas             int32
	PVCName              string
	InitContainerName    string
	InitContainerCommand []string
	ReleaseIdentifier    string
}

func CreateOrUpdateDeployment(ctx context.Context, clientset *kubernetes.Clientset, params CreateOrUpdateDeploymentParams) error {
	deploymentsClient := clientset.AppsV1().Deployments(params.Namespace)
	replicas := cmp.Or(params.Replicas, int32(1))
	namespace := cmp.Or(params.Namespace, "default")
	envVarDeployTimestamp := "K8RUN_DEPLOY_TIMESTAMP"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: namespace,
			Labels: map[string]string{
				LabelNameCreatedBy:         LabelValueCreatedBy,
				LabelNameReleaseIdentifier: params.ReleaseIdentifier,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": params.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                      params.Name,
						LabelNameCreatedBy:         LabelValueCreatedBy,
						LabelNameReleaseIdentifier: params.ReleaseIdentifier,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:    params.InitContainerName,
							Image:   "busybox",
							Command: params.InitContainerCommand,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "app",
									MountPath: params.CopyFolderTo,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envVarDeployTimestamp,
									Value: time.Now().Format(time.RFC3339),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:       params.Name,
							Image:      params.Image,
							Args:       params.Entrypoint,
							WorkingDir: params.CopyFolderTo,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: params.ContainerPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "app",
									MountPath: params.CopyFolderTo,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envVarDeployTimestamp,
									Value: time.Now().Format(time.RFC3339),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "app",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: params.PVCName,
								},
							},
						},
					},
				},
			},
		},
	}

	// Try to create or update the deployment
	existentDeployment, err := deploymentsClient.Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		_, err = deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create deployment: %w", err)
		}
		slog.With("name", params.Name, "namespace", namespace).Info("Deployment created")
	} else {
		if existentDeployment.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("deployment already exists but it has not been created by k8run")
		}

		_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update deployment: %w", err)
		}
		slog.With("name", params.Name, "namespace", namespace).Info("Deployment updated")
	}

	return nil
}
