package k8s

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"cmp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CreateOrUpdateDeploymentParams struct {
	Name                 string
	Namespace            string
	Entrypoint           []string
	ContainerPort        int32
	Image                string
	CopyTo               string
	Replicas             int32
	PVCName              string
	InitContainerName    string
	InitContainerCommand []string
	ReleaseIdentifier    string
}

func CreateOrUpdateDeployment(ctx context.Context, clientset *kubernetes.Clientset, params CreateOrUpdateDeploymentParams) error {
	deploymentsClient := clientset.AppsV1().Deployments(params.Namespace)
	replicas := cmp.Or(params.Replicas, int32(1))
	envVarDeployTimestamp := "K8RUN_DEPLOY_TIMESTAMP"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: params.Namespace,
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
									MountPath: params.CopyTo,
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
							WorkingDir: params.CopyTo,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: params.ContainerPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "app",
									MountPath: params.CopyTo,
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
		slog.With("name", params.Name, "namespace", params.Namespace).Info("Deployment created")
	} else {
		if existentDeployment.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("deployment already exists but it has not been created by k8run")
		}

		_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update deployment: %w", err)
		}
		slog.With("name", params.Name, "namespace", params.Namespace).Info("Deployment updated")
	}

	return nil
}

type DeleteDeploymentParams struct {
	Name      string
	Namespace string
}

func DeleteDeployment(ctx context.Context, clientset *kubernetes.Clientset, params DeleteDeploymentParams) error {
	deploymentsClient := clientset.AppsV1().Deployments(params.Namespace)

	existentDeployment, err := GetDeployment(ctx, clientset, GetParams{
		Name:      params.Name,
		Namespace: params.Namespace,
	})
	if err != nil {
		return err
	}

	if existentDeployment.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
		return fmt.Errorf("deployment already exists but it has not been created by k8run")
	}

	err = deploymentsClient.Delete(ctx, params.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	slog.With("name", params.Name, "namespace", params.Namespace).Info("Deployment marked for deletion")
	return nil
}

func GetDeployment(ctx context.Context, clientset *kubernetes.Clientset, params GetParams) (*appsv1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(params.Namespace)
	existentDeployment, err := deploymentsClient.Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("deployment %q not found in namespace %q: %w", params.Name, params.Namespace, ErrResourceNotFound)
		}

		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return existentDeployment, nil
}

type WaitForDeploymentToBeReadyParams struct {
	Name              string
	Namespace         string
	ReleaseIdentifier string
}

func WaitForDeploymentToBeReady(ctx context.Context, clientset *kubernetes.Clientset, params WaitForDeploymentToBeReadyParams) error {
	for {
		deployment, err := GetDeployment(ctx, clientset, GetParams{
			Name:      params.Name,
			Namespace: params.Namespace,
		})
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}

		if deployment.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("deployment already exists but it has not been created by k8run")
		}

		if deployment.Labels[LabelNameReleaseIdentifier] == params.ReleaseIdentifier && deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
			break
		}

		slog.With("name", params.Name, "namespace", params.Namespace).Info("Waiting for deployment to be ready...")
		time.Sleep(2 * time.Second)
	}

	return nil
}
