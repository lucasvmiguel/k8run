package k8s_test

import (
	"context"
	"testing"

	"github.com/lucasvmiguel/k8run/internal/k8s"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateOrUpdateIngress_CreateNew(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	params := k8s.CreateOrUpdateIngressParams{
		Name:         "test-ingress",
		Namespace:    "default",
		IngressClass: nil,
		IngressHost:  "example.com",
		Port:         80,
	}

	err := k8s.CreateOrUpdateIngress(context.Background(), clientset, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ingress, err := clientset.NetworkingV1().Ingresses(params.Namespace).Get(context.Background(), params.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get created ingress: %v", err)
	}

	if ingress.Spec.Rules[0].Host != params.IngressHost {
		t.Errorf("expected host %s, got %s", params.IngressHost, ingress.Spec.Rules[0].Host)
	}
}

func TestCreateOrUpdateIngress_UpdateExisting(t *testing.T) {
	clientset := fake.NewSimpleClientset(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: k8s.LabelValueCreatedBy,
			},
		},
	})

	params := k8s.CreateOrUpdateIngressParams{
		Name:         "test-ingress",
		Namespace:    "default",
		IngressClass: nil,
		IngressHost:  "new.example.com",
		Port:         443,
	}

	err := k8s.CreateOrUpdateIngress(context.Background(), clientset, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ingress, err := clientset.NetworkingV1().Ingresses(params.Namespace).Get(context.Background(), params.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get updated ingress: %v", err)
	}

	if ingress.Spec.Rules[0].Host != params.IngressHost {
		t.Errorf("expected host %s, got %s", params.IngressHost, ingress.Spec.Rules[0].Host)
	}
}

func TestDeleteIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
			Labels: map[string]string{
				k8s.LabelNameCreatedBy: k8s.LabelValueCreatedBy,
			},
		},
	})

	err := k8s.DeleteIngress(context.Background(), clientset, k8s.DeleteIngressParams{
		Name:      "test-ingress",
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = clientset.NetworkingV1().Ingresses("default").Get(context.Background(), "test-ingress", metav1.GetOptions{})
	if err == nil {
		t.Fatalf("expected ingress to be deleted, but it still exists")
	}
}

func TestDeleteIngress_NotCreatedByK8Run(t *testing.T) {
	clientset := fake.NewSimpleClientset(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
			Labels: map[string]string{
				"creator": "other-controller",
			},
		},
	})

	err := k8s.DeleteIngress(context.Background(), clientset, k8s.DeleteIngressParams{
		Name:      "test-ingress",
		Namespace: "default",
	})
	if err == nil {
		t.Fatalf("expected error due to incorrect creator label")
	}
}

func TestGetIngress(t *testing.T) {
	clientset := fake.NewSimpleClientset(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
		},
	})

	ingress, err := k8s.GetIngress(context.Background(), clientset, k8s.GetParams{
		Name:      "test-ingress",
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ingress.Name != "test-ingress" {
		t.Errorf("expected ingress name 'test-ingress', got %s", ingress.Name)
	}
}

func TestGetIngress_NotFound(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	_, err := k8s.GetIngress(context.Background(), clientset, k8s.GetParams{
		Name:      "nonexistent-ingress",
		Namespace: "default",
	})
	if err == nil {
		t.Fatalf("expected error for non-existent ingress, got nil")
	}
}
