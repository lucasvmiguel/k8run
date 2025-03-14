package k8s_test

import (
	"context"
	"testing"

	"github.com/lucasvmiguel/k8run/internal/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateOrUpdateService_CreateNew(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	params := k8s.CreateOrUpdateServiceParams{
		Name:              "test-service",
		Namespace:         "default",
		Port:              80,
		ContainerPort:     8080,
		ReleaseIdentifier: "v1",
	}

	err := k8s.CreateOrUpdateService(context.Background(), clientset, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	service, err := clientset.CoreV1().Services(params.Namespace).Get(context.Background(), params.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get created service: %v", err)
	}

	if service.Spec.Ports[0].Port != params.Port {
		t.Errorf("expected port %d, got %d", params.Port, service.Spec.Ports[0].Port)
	}

	if service.Spec.Ports[0].TargetPort.IntVal != int32(params.ContainerPort) {
		t.Errorf("expected container port %d, got %d", params.ContainerPort, service.Spec.Ports[0].TargetPort.IntVal)
	}

	if service.Labels[k8s.LabelNameReleaseIdentifier] != params.ReleaseIdentifier {
		t.Errorf("expected release identifier %s, got %s", params.ReleaseIdentifier, service.Labels[k8s.LabelNameReleaseIdentifier])
	}
}

func TestCreateOrUpdateService_UpdateExisting(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: k8s.LabelValueCreatedBy,
			},
		},
	})

	params := k8s.CreateOrUpdateServiceParams{
		Name:              "test-service",
		Namespace:         "default",
		Port:              443,
		ContainerPort:     8443,
		ReleaseIdentifier: "v2",
	}

	err := k8s.CreateOrUpdateService(context.Background(), clientset, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	service, err := clientset.CoreV1().Services(params.Namespace).Get(context.Background(), params.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get updated service: %v", err)
	}

	if service.Spec.Ports[0].Port != params.Port {
		t.Errorf("expected port %d, got %d", params.Port, service.Spec.Ports[0].Port)
	}

	if service.Spec.Ports[0].TargetPort.IntVal != int32(params.ContainerPort) {
		t.Errorf("expected container port %d, got %d", params.ContainerPort, service.Spec.Ports[0].TargetPort.IntVal)
	}

	if service.Labels[k8s.LabelNameReleaseIdentifier] != params.ReleaseIdentifier {
		t.Errorf("expected release identifier %s, got %s", params.ReleaseIdentifier, service.Labels[k8s.LabelNameReleaseIdentifier])
	}
}

func TestCreateOrUpdateService_ConflictWithNonK8Run(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: "other-controller",
			},
		},
	})

	params := k8s.CreateOrUpdateServiceParams{
		Name:              "test-service",
		Namespace:         "default",
		Port:              443,
		ContainerPort:     8443,
		ReleaseIdentifier: "v2",
	}

	err := k8s.CreateOrUpdateService(context.Background(), clientset, params)
	if err == nil {
		t.Fatalf("expected error due to conflict with non-k8run service")
	}
}

func TestDeleteService_Success(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: k8s.LabelValueCreatedBy,
			},
		},
	})

	err := k8s.DeleteService(context.Background(), clientset, k8s.DeleteServiceParams{
		Name:      "test-service",
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = clientset.CoreV1().Services("default").Get(context.Background(), "test-service", metav1.GetOptions{})
	if err == nil {
		t.Fatalf("expected service to be deleted, but it still exists")
	}
}

func TestDeleteService_ConflictWithNonK8Run(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: "other-controller",
			},
		},
	})

	err := k8s.DeleteService(context.Background(), clientset, k8s.DeleteServiceParams{
		Name:      "test-service",
		Namespace: "default",
	})
	if err == nil {
		t.Fatalf("expected error due to conflict with non-k8run service")
	}
}

func TestGetService_Success(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
	})

	service, err := k8s.GetService(context.Background(), clientset, k8s.GetParams{
		Name:      "test-service",
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if service.Name != "test-service" {
		t.Errorf("expected service name 'test-service', got %s", service.Name)
	}
}

func TestGetService_NotFound(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	_, err := k8s.GetService(context.Background(), clientset, k8s.GetParams{
		Name:      "nonexistent-service",
		Namespace: "default",
	})
	if err == nil {
		t.Fatalf("expected error for non-existent service, got nil")
	}
}
