package k8s

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type CreateOrUpdateServiceParams struct {
	Name              string
	Namespace         string
	Port              int32
	ContainerPort     int32
	ReleaseIdentifier string
}

func CreateOrUpdateService(ctx context.Context, clientset *kubernetes.Clientset, params CreateOrUpdateServiceParams) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: params.Namespace,
			Labels: map[string]string{
				LabelNameCreatedBy:         LabelValueCreatedBy,
				LabelNameReleaseIdentifier: params.ReleaseIdentifier,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": params.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       params.Port,
					TargetPort: intstr.FromInt(int(params.ContainerPort)),
				},
			},
		},
	}

	existingService, err := clientset.CoreV1().Services(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		_, err = clientset.CoreV1().Services(params.Namespace).Create(ctx, service, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service: %v", err)
		}
		slog.With("name", params.Name, "namespace", params.Namespace).Info("Service created")
	} else {
		if existingService.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("service already exists but it has not been created by k8run")
		}

		service.ResourceVersion = existingService.ResourceVersion
		_, err = clientset.CoreV1().Services(params.Namespace).Update(ctx, service, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update service: %v", err)
		}
		slog.With("name", params.Name, "namespace", params.Namespace).Info("Service updated")
	}

	return nil
}

type DeleteServiceParams struct {
	Name      string
	Namespace string
}

func DeleteService(ctx context.Context, clientset *kubernetes.Clientset, params DeleteServiceParams) error {
	servicesClient := clientset.CoreV1().Services(params.Namespace)

	existentService, err := GetService(ctx, clientset, GetParams{
		Name:      params.Name,
		Namespace: params.Namespace,
	})
	if err != nil {
		return err
	}

	if existentService.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
		return fmt.Errorf("service already exists but it has not been created by k8run")
	}

	err = servicesClient.Delete(ctx, params.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	slog.With("name", params.Name, "namespace", params.Namespace).Info("Service marked for deletion")
	return nil
}

func GetService(ctx context.Context, clientset *kubernetes.Clientset, params GetParams) (*corev1.Service, error) {
	servicesClient := clientset.CoreV1().Services(params.Namespace)
	service, err := servicesClient.Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("service %q not found in namespace %q: %w", params.Name, params.Namespace, ErrResourceNotFound)
		}

		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return service, nil
}
