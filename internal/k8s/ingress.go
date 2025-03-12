package k8s

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"

	networkingv1 "k8s.io/api/networking/v1"
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
	namespace := cmp.Or(params.Namespace, "default")

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: namespace,
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

	existingIngress, err := clientset.NetworkingV1().Ingresses(namespace).Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		_, err = clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ingress: %v", err)
		}
		slog.With("name", params.Name, "namespace", namespace).Info("Ingress created")
	} else {
		if existingIngress.Labels[LabelNameCreatedBy] != LabelValueCreatedBy {
			return fmt.Errorf("ingress already exists but it has not been created by k8run")
		}

		ingress.ResourceVersion = existingIngress.ResourceVersion
		_, err = clientset.NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update ingress: %v", err)
		}
		slog.With("name", params.Name, "namespace", namespace).Info("Ingress updated")
	}

	return nil
}
