package k8s_test

import (
	"context"
	"testing"
	"time"

	"github.com/lucasvmiguel/k8run/internal/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestWaitForRunningInitContainer_Success(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameReleaseIdentifier: "release-123",
			},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "test-init-container",
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.Time{Time: time.Now()},
						},
					},
				},
			},
		},
	})

	params := k8s.WaitForRunningInitContainerParams{
		Namespace:         "default",
		Name:              "test-pod",
		InitContainerName: "test-init-container",
		ReleaseIdentifier: "release-123",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pod, err := k8s.WaitForRunningInitContainer(ctx, clientset, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if pod == nil || pod.Name != "test-pod" {
		t.Fatalf("expected pod name 'test-pod', got %v", pod)
	}
}

func TestWaitForRunningInitContainer_ContextCancelled(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	params := k8s.WaitForRunningInitContainerParams{
		Namespace:         "default",
		Name:              "test-pod",
		InitContainerName: "test-init-container",
		ReleaseIdentifier: "release-123",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel context

	_, err := k8s.WaitForRunningInitContainer(ctx, clientset, params)
	if err == nil || err.Error() != "context cancelled while waiting for init container to be running" {
		t.Fatalf("expected context cancellation error, got %v", err)
	}
}

func TestWaitForRunningInitContainer_NoMatchingPods(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	params := k8s.WaitForRunningInitContainerParams{
		Namespace:         "default",
		Name:              "test-pod",
		InitContainerName: "test-init-container",
		ReleaseIdentifier: "release-123",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := k8s.WaitForRunningInitContainer(ctx, clientset, params)
	if err == nil {
		t.Fatalf("expected error due to no matching pods, got nil")
	}
}

func TestWaitForRunningInitContainer_InitContainerNotRunning(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameReleaseIdentifier: "release-123",
			},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "test-init-container",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "ContainerCreating",
						},
					},
				},
			},
		},
	})

	params := k8s.WaitForRunningInitContainerParams{
		Namespace:         "default",
		Name:              "test-pod",
		InitContainerName: "test-init-container",
		ReleaseIdentifier: "release-123",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := k8s.WaitForRunningInitContainer(ctx, clientset, params)
	if err == nil {
		t.Fatalf("expected timeout or init container not running error, got nil")
	}
}
