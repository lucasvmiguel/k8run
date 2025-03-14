package k8s_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lucasvmiguel/k8run/internal/k8s"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreatePVCIfNotExists_CreateNew(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	params := k8s.CreatePVCIfNotExistsParams{
		Name:      "test-pvc",
		Namespace: "default",
	}

	err := k8s.CreatePVCIfNotExists(context.Background(), clientset, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	pvc, err := clientset.CoreV1().PersistentVolumeClaims(params.Namespace).Get(context.Background(), params.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get created PVC: %v", err)
	}

	if pvc.Labels[k8s.LabelNameCreatedBy] != k8s.LabelValueCreatedBy {
		t.Errorf("expected label %s=%s, got %v", k8s.LabelNameCreatedBy, k8s.LabelValueCreatedBy, pvc.Labels)
	}

	expectedStorage := resource.MustParse("1Gi")
	if pvc.Spec.Resources.Requests[corev1.ResourceStorage] != expectedStorage {
		t.Errorf("expected storage size %s, got %s", expectedStorage.String(), pvc.Spec.Resources.Requests[corev1.ResourceStorage].ToUnstructured())
	}
}

func TestCreatePVCIfNotExists_AlreadyExistsByK8Run(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: k8s.LabelValueCreatedBy,
			},
		},
	})

	params := k8s.CreatePVCIfNotExistsParams{
		Name:      "test-pvc",
		Namespace: "default",
	}

	err := k8s.CreatePVCIfNotExists(context.Background(), clientset, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreatePVCIfNotExists_ConflictWithNonK8Run(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: "other-controller",
			},
		},
	})

	params := k8s.CreatePVCIfNotExistsParams{
		Name:      "test-pvc",
		Namespace: "default",
	}

	err := k8s.CreatePVCIfNotExists(context.Background(), clientset, params)
	if err == nil {
		t.Fatalf("expected error due to conflicting label, got nil")
	}
}

func TestDeletePVC_Success(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: k8s.LabelValueCreatedBy,
			},
		},
	})

	err := k8s.DeletePVC(context.Background(), clientset, k8s.DeletePVCParams{
		Name:      "test-pvc",
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = clientset.CoreV1().PersistentVolumeClaims("default").Get(context.Background(), "test-pvc", metav1.GetOptions{})
	if err == nil {
		t.Fatalf("expected PVC to be deleted, but it still exists")
	}
	if !k8serrors.IsNotFound(err) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestDeletePVC_ConflictWithNonK8Run(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: "other-controller",
			},
		},
	})

	err := k8s.DeletePVC(context.Background(), clientset, k8s.DeletePVCParams{
		Name:      "test-pvc",
		Namespace: "default",
	})
	if err == nil {
		t.Fatalf("expected error due to conflicting label, got nil")
	}
}

func TestGetPVC_Success(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
		},
	})

	pvc, err := k8s.GetPVC(context.Background(), clientset, k8s.GetParams{
		Name:      "test-pvc",
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if pvc.Name != "test-pvc" {
		t.Errorf("expected PVC name 'test-pvc', got %s", pvc.Name)
	}
}

func TestGetPVC_NotFound(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	_, err := k8s.GetPVC(context.Background(), clientset, k8s.GetParams{
		Name:      "nonexistent-pvc",
		Namespace: "default",
	})
	if err == nil {
		t.Fatalf("expected error for non-existent PVC, got nil")
	}
	if !errors.Is(err, k8s.ErrResourceNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
