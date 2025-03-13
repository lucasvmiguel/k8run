package command

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/lucasvmiguel/k8run/internal/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type NewDestroyCommandParams struct {
	Name      string
	Namespace string
	Timeout   time.Duration
}

type DestroyCommand struct {
	Name      string
	Namespace string
	Timeout   time.Duration
}

func NewDestroyCommand(params NewDestroyCommandParams) *DestroyCommand {
	return &DestroyCommand{
		Name:      params.Name,
		Namespace: params.Namespace,
		Timeout:   params.Timeout,
	}
}

func (c *DestroyCommand) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("Name is required")
	}

	if c.Timeout < 10*time.Second {
		return fmt.Errorf("Timeout must be greater than 10s")
	}
	return nil
}

func (c *DestroyCommand) Run(ctx context.Context) error {
	slog.Info("Starting destroying...")
	c.Namespace = cmp.Or(c.Namespace, "default")

	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return fmt.Errorf("Failed to build k8s config: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Failed to create k8s clientset: %s", err)
	}

	wg := sync.WaitGroup{}
	deletingDeployment := false
	deletingPVC := false
	deletingIngress := false
	deletingService := false

	err = k8s.DeleteDeployment(ctx, clientset, k8s.DeleteDeploymentParams{
		Name:      c.Name,
		Namespace: c.Namespace,
	})
	if err != nil {
		if errors.Is(err, k8s.ErrResourceNotFound) {
			slog.With("name", c.Name, "namespace", c.Namespace).Info("Deployment not found")
		} else {
			return fmt.Errorf("Failed to delete deployment: %s", err)
		}
	} else {
		deletingDeployment = true
		wg.Add(1)
	}

	pvcName := pvcName(c.Name)
	err = k8s.DeletePVC(ctx, clientset, k8s.DeletePVCParams{
		Name:      pvcName,
		Namespace: c.Namespace,
	})
	if err != nil {
		if errors.Is(err, k8s.ErrResourceNotFound) {
			slog.With("name", c.Name, "namespace", c.Namespace).Info("PVC not found")
		} else {
			return fmt.Errorf("Failed to delete PVC: %s", err)
		}
	} else {
		deletingPVC = true
		wg.Add(1)
	}

	err = k8s.DeleteService(ctx, clientset, k8s.DeleteServiceParams{
		Name:      c.Name,
		Namespace: c.Namespace,
	})
	if err != nil {
		if errors.Is(err, k8s.ErrResourceNotFound) {
			slog.With("name", c.Name, "namespace", c.Namespace).Info("Service not found")
		} else {
			return fmt.Errorf("Failed to delete service: %s", err)
		}
	} else {
		deletingService = true
		wg.Add(1)
	}

	err = k8s.DeleteIngress(ctx, clientset, k8s.DeleteIngressParams{
		Name:      c.Name,
		Namespace: c.Namespace,
	})
	if err != nil {
		if errors.Is(err, k8s.ErrResourceNotFound) {
			slog.With("name", c.Name, "namespace", c.Namespace).Info("Ingress not found")
		} else {
			return fmt.Errorf("Failed to delete ingress: %s", err)
		}
	} else {
		deletingIngress = true
		wg.Add(1)
	}

	if deletingDeployment {
		go func() {
			defer wg.Done()
			for {
				_, err := k8s.GetDeployment(ctx, clientset, k8s.GetParams{
					Name:      c.Name,
					Namespace: c.Namespace,
				})
				if errors.Is(err, k8s.ErrResourceNotFound) {
					slog.With("name", c.Name, "namespace", c.Namespace).Info("Deployment deleted")
					return
				}
				time.Sleep(2 * time.Second)
				slog.With("name", c.Name, "namespace", c.Namespace).Info("Waiting for deployment deletion...")
			}
		}()
	}

	if deletingPVC {
		go func() {
			defer wg.Done()
			for {
				_, err := k8s.GetPVC(ctx, clientset, k8s.GetParams{
					Name:      pvcName,
					Namespace: c.Namespace,
				})
				if errors.Is(err, k8s.ErrResourceNotFound) {
					slog.With("name", pvcName, "namespace", c.Namespace).Info("PVC deleted")
					return
				}
				time.Sleep(2 * time.Second)
				slog.With("name", pvcName, "namespace", c.Namespace).Info("Waiting for PVC deletion...")
			}
		}()
	}

	if deletingService {
		go func() {
			defer wg.Done()
			for {
				_, err := k8s.GetService(ctx, clientset, k8s.GetParams{
					Name:      c.Name,
					Namespace: c.Namespace,
				})
				if errors.Is(err, k8s.ErrResourceNotFound) {
					slog.With("name", c.Name, "namespace", c.Namespace).Info("Service deleted")
					return
				}
				time.Sleep(2 * time.Second)
				slog.With("name", c.Name, "namespace", c.Namespace).Info("Waiting for service deletion...")
			}
		}()
	}

	if deletingIngress {
		go func() {
			defer wg.Done()
			for {
				_, err := k8s.GetIngress(ctx, clientset, k8s.GetParams{
					Name:      c.Name,
					Namespace: c.Namespace,
				})
				if errors.Is(err, k8s.ErrResourceNotFound) {
					slog.With("name", c.Name, "namespace", c.Namespace).Info("Ingress deleted")
					return
				}
				time.Sleep(2 * time.Second)
				slog.With("name", c.Name, "namespace", c.Namespace).Info("Waiting for ingress deletion...")
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		return fmt.Errorf("Timeout while waiting for resource deletion")
	}

	slog.Info("Destroy finished!")

	return nil
}
