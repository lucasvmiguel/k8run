package k8s_test

import (
	"context"
	"testing"

	"github.com/lucasvmiguel/k8run/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateOrUpdateDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	params := k8s.CreateOrUpdateDeploymentParams{
		Name:                 "test-deployment",
		Namespace:            "default",
		Entrypoint:           []string{"./app"},
		ContainerPort:        8080,
		Image:                "test-image",
		CopyTo:               "/app",
		Replicas:             2,
		PVCName:              "test-pvc",
		InitContainerName:    "init-container",
		InitContainerCommand: []string{"sh", "-c", "echo 'Init'"},
		ReleaseIdentifier:    "test-release",
	}

	err := k8s.CreateOrUpdateDeployment(context.TODO(), clientset, params)
	if err != nil {
		t.Fatalf("failed to create deployment: %v", err)
	}

	// Check that the deployment was created
	deployment, err := clientset.AppsV1().Deployments(params.Namespace).Get(context.TODO(), params.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get created deployment: %v", err)
	}

	if *deployment.Spec.Replicas != params.Replicas {
		t.Errorf("expected replicas to be %d, got %d", params.Replicas, *deployment.Spec.Replicas)
	}
}

func TestDeleteDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: k8s.LabelValueCreatedBy,
			},
		},
	})

	params := k8s.DeleteDeploymentParams{
		Name:      "test-deployment",
		Namespace: "default",
	}

	err := k8s.DeleteDeployment(context.TODO(), clientset, params)
	if err != nil {
		t.Fatalf("failed to delete deployment: %v", err)
	}

	// Verify that the deployment is deleted
	_, err = clientset.AppsV1().Deployments(params.Namespace).Get(context.TODO(), params.Name, metav1.GetOptions{})
	if err == nil {
		t.Fatalf("deployment was not deleted")
	}
}

func TestGetDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: k8s.LabelValueCreatedBy,
			},
		},
	})

	params := k8s.GetParams{
		Name:      "test-deployment",
		Namespace: "default",
	}

	deployment, err := k8s.GetDeployment(context.TODO(), clientset, params)
	if err != nil {
		t.Fatalf("failed to get deployment: %v", err)
	}

	if deployment.Name != params.Name {
		t.Errorf("expected deployment name to be %q, got %q", params.Name, deployment.Name)
	}
}

func TestWaitForDeploymentToBeReady(t *testing.T) {
	clientset := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy:         k8s.LabelValueCreatedBy,
				k8s.LabelNameReleaseIdentifier: "test-release",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 1,
		},
	})

	params := k8s.WaitForDeploymentToBeReadyParams{
		Name:              "test-deployment",
		Namespace:         "default",
		ReleaseIdentifier: "test-release",
	}

	err := k8s.WaitForDeploymentToBeReady(context.TODO(), clientset, params)
	if err != nil {
		t.Fatalf("deployment not ready: %v", err)
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
