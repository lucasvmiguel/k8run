package k8s

import (
	"context"
	"fmt"
	"log/slog"

	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CreateOrUpdateIngressParams struct {
	Name         string
	Namespace    string
	IngressClass *string
	IngressHost  string
	Port         int32
}

func CreateOrUpdateIngress(ctx context.Context, clientset *kubernetes.Clientset, params CreateOrUpdateIngressParams) error {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: params.Namespace,
			Labels: map[string]string{
				LabelNameCreatedBy: LabelValueCreatedBy,
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: params.IngressClass,
			Rules: []networkingv1.IngressRule{
				{
					Host: params.IngressHost,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: func() *networkingv1.PathType { t := networkingv1.PathTypePrefix; return &t }(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: params.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: params.Port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	existingIngress, err := clientset.NetworkingV1().Ingresses(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		_, err = clientset.NetworkingV1().Ingresses(params.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ingress: %v", err)
		}
		slog.With("name", params.Name, "namespace", params.Namespace).Info("Ingress created")
	} else {
		if existingIngress.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("ingress already exists but it has not been created by k8run")
		}

		ingress.ResourceVersion = existingIngress.ResourceVersion
		_, err = clientset.NetworkingV1().Ingresses(params.Namespace).Update(ctx, ingress, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update ingress: %v", err)
		}
		slog.With("name", params.Name, "namespace", params.Namespace).Info("Ingress updated")
	}

	return nil
}

type DeleteIngressParams struct {
	Name      string
	Namespace string
}

func DeleteIngress(ctx context.Context, clientset *kubernetes.Clientset, params DeleteIngressParams) error {
	ingressesClient := clientset.NetworkingV1().Ingresses(params.Namespace)

	existentIngress, err := GetIngress(ctx, clientset, GetParams{
		Name:      params.Name,
		Namespace: params.Namespace,
	})
	if err != nil {
		return err
	}

	if existentIngress.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
		return fmt.Errorf("ingress already exists but it has not been created by k8run")
	}

	err = ingressesClient.Delete(ctx, params.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete ingress: %w", err)
	}

	slog.With("name", params.Name, "namespace", params.Namespace).Info("Ingress marked for deletion")
	return nil
}

func GetIngress(ctx context.Context, clientset *kubernetes.Clientset, params GetParams) (*networkingv1.Ingress, error) {
	ingressesClient := clientset.NetworkingV1().Ingresses(params.Namespace)
	existentIngress, err := ingressesClient.Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("ingress %q not found in namespace %q: %w", params.Name, params.Namespace, ErrResourceNotFound)
		}

		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}

	return existentIngress, nil
}
