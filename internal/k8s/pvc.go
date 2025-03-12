package k8s

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CreatePVCIfNotExistsParams struct {
	Name      string
	Namespace string
}

func CreatePVCIfNotExists(ctx context.Context, clientset *kubernetes.Clientset, params CreatePVCIfNotExistsParams) error {
	namespace := cmp.Or(params.Namespace, "default")

	pvcClient := clientset.CoreV1().PersistentVolumeClaims(namespace)
	pvc, err := pvcClient.Get(ctx, params.Name, metav1.GetOptions{})
	if err == nil {
		if pvc.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("PVC with the same name already exists but it was not created by k8run")
		}

		slog.With("name", params.Name, "namespace", namespace).Info("PVC already exists")
		return nil
	}

	pvcParams := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: params.Name,
			Labels: map[string]string{
				LabelNameCreatedBy: LabelValueCreatedBy,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	_, err = pvcClient.Create(ctx, pvcParams, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create PVC: %w", err)
	}

	slog.With("name", params.Name, "namespace", namespace).Info("PVC created")
	return nil
}
