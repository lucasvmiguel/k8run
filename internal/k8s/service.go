package k8s

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
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
	namespace := cmp.Or(params.Namespace, "default")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: namespace,
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

	existingService, err := clientset.CoreV1().Services(namespace).Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		_, err = clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service: %v", err)
		}
		slog.With("name", params.Name, "namespace", namespace).Info("Service created")
	} else {
		if existingService.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("service already exists but it has not been created by k8run")
		}

		service.ResourceVersion = existingService.ResourceVersion
		_, err = clientset.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update service: %v", err)
		}
		slog.With("name", params.Name, "namespace", namespace).Info("Service updated")
	}

	return nil
}
