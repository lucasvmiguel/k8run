package k8s

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreatePVCIfNotExistsParams represents the parameters to create a PVC if it does not exist.
type CreatePVCIfNotExistsParams struct {
	Name      string
	Namespace string
}

// CreatePVCIfNotExists creates a PVC if it does not exist in the given namespace.
func CreatePVCIfNotExists(ctx context.Context, clientset kubernetes.Interface, params CreatePVCIfNotExistsParams) error {
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(params.Namespace)
	pvc, err := pvcClient.Get(ctx, params.Name, metav1.GetOptions{})
	if err == nil {
		if pvc.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("PVC with the same name already exists but it was not created by k8run")
		}

		slog.With("name", params.Name, "namespace", params.Namespace).Info("PVC already exists")
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

	slog.With("name", params.Name, "namespace", params.Namespace).Info("PVC created")
	return nil
}

// DeletePVCParams represents the parameters to delete a PVC.
type DeletePVCParams struct {
	Name      string
	Namespace string
}

// DeletePVC deletes a PVC in the given namespace.
func DeletePVC(ctx context.Context, clientset kubernetes.Interface, params DeletePVCParams) error {
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(params.Namespace)

	pvc, err := GetPVC(ctx, clientset, GetParams{
		Name:      params.Name,
		Namespace: params.Namespace,
	})
	if err != nil {
		return err
	}

	if pvc.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
		return fmt.Errorf("PVC already exists but it has not been created by k8run")
	}

	err = pvcClient.Delete(ctx, params.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete PVC: %w", err)
	}

	slog.With("name", params.Name, "namespace", params.Namespace).Info("PVC marked for deletion")
	return nil
}

// GetPVC retrieves a PVC in the given namespace.
func GetPVC(ctx context.Context, clientset kubernetes.Interface, params GetParams) (*corev1.PersistentVolumeClaim, error) {
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(params.Namespace)
	pvc, err := pvcClient.Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("pvc %q not found in namespace %q: %w", params.Name, params.Namespace, ErrResourceNotFound)
		}

		return nil, fmt.Errorf("failed to get PVC: %w", err)
	}

	return pvc, nil
}
