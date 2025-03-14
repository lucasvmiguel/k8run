package command

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lucasvmiguel/k8run/internal/k8s"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// NewDeploymentCommandParams represents the parameters to create a new deployment command.
type NewDeploymentCommandParams struct {
	Name          string
	Entrypoint    []string
	Copy          string
	ContainerPort int64
	Port          int64
	Service       bool
	Ingress       bool
	IngressHost   string
	IngressClass  string
	Namespace     string
	Image         string
	Replicas      int32
	Timeout       time.Duration
}

// DeploymentCommand represents a command to deploy an application and its related resources in a Kubernetes cluster.
type DeploymentCommand struct {
	Name          string
	Entrypoint    []string
	Copy          string
	ContainerPort int64
	Port          int64
	Service       bool
	Ingress       bool
	IngressHost   string
	IngressClass  string
	Namespace     string
	Image         string
	Replicas      int32
	Timeout       time.Duration
}

// NewDeploymentCommand creates a new deployment command.
func NewDeploymentCommand(params NewDeploymentCommandParams) *DeploymentCommand {
	return &DeploymentCommand{
		Name:          params.Name,
		Entrypoint:    params.Entrypoint,
		Copy:          params.Copy,
		ContainerPort: params.ContainerPort,
		Port:          params.Port,
		Service:       params.Service,
		Ingress:       params.Ingress,
		IngressHost:   params.IngressHost,
		IngressClass:  params.IngressClass,
		Namespace:     params.Namespace,
		Image:         params.Image,
		Replicas:      params.Replicas,
		Timeout:       params.Timeout,
	}
}

// Validate validates the parameters of the deployment command.
func (c *DeploymentCommand) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("Name is required")
	}
	if c.Image == "" {
		return fmt.Errorf("Image is required")
	}
	if c.Copy == "" {
		return fmt.Errorf("Copy is required")
	}
	if c.Replicas < 1 {
		return fmt.Errorf("Replicas must be greater than 0")
	}
	if c.Timeout < 10*time.Second {
		return fmt.Errorf("Timeout must be greater than 10s")
	}
	if c.Service {
		if c.Port < 0 {
			return fmt.Errorf("Port must be greater than 0")
		}
		if c.ContainerPort < 0 {
			return fmt.Errorf("ContainerPort must be greater than 0")
		}
	}
	if c.Ingress {
		if c.Port < 0 {
			return fmt.Errorf("Port must be greater than 0")
		}

		if c.IngressHost == "" {
			return fmt.Errorf("IngressHost is required")
		}

		if c.IngressClass == "" {
			return fmt.Errorf("IngressClass is required")
		}
	}
	return nil
}

// Run runs the deployment command.
func (c *DeploymentCommand) Run(ctx context.Context) error {
	slog.Info("Starting deployment...")
	c.Namespace = cmp.Or(c.Namespace, "default")

	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	copyTo := "/app"
	initContainerName := "wait-to-copy-app"

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return fmt.Errorf("Failed to build k8s config: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Failed to create k8s clientset: %s", err)
	}

	pvcName := pvcName(c.Name)
	err = k8s.CreatePVCIfNotExists(ctx, clientset, k8s.CreatePVCIfNotExistsParams{
		Name:      pvcName,
		Namespace: c.Namespace,
	})
	if err != nil {
		return fmt.Errorf("Failed to create PVC: %s", err)
	}

	releaseIdentifier := rand.String(10)
	err = k8s.CreateOrUpdateDeployment(ctx, clientset, k8s.CreateOrUpdateDeploymentParams{
		Name:              c.Name,
		Namespace:         c.Namespace,
		Entrypoint:        c.Entrypoint,
		ContainerPort:     int32(c.ContainerPort),
		Image:             c.Image,
		CopyTo:            copyTo,
		Replicas:          c.Replicas,
		PVCName:           pvcName,
		InitContainerName: initContainerName,
		ReleaseIdentifier: releaseIdentifier,
		InitContainerCommand: []string{
			"sh", "-c", fmt.Sprintf(
				`rm -rf %s/* && until [ -n "$(ls -A %s)" ]; do echo "Waiting for folder to be non-empty"; sleep 5; done; sleep 2; exit 0`,
				copyTo, copyTo),
		},
	})
	if err != nil {
		return fmt.Errorf("Failed to create or update deployment: %s", err)
	}

	pod, err := k8s.WaitForRunningInitContainer(ctx, clientset, k8s.WaitForRunningInitContainerParams{
		Namespace:         c.Namespace,
		Name:              c.Name,
		InitContainerName: initContainerName,
		ReleaseIdentifier: releaseIdentifier,
	})
	if err != nil {
		return fmt.Errorf("Failed to wait for init container: %s", err)
	}

	err = k8s.CopyToPod(k8s.CopyToPodParams{
		LocalPath:         c.Copy,
		PodName:           pod.Name,
		ContainerPath:     copyTo,
		InitContainerName: initContainerName,
		Namespace:         c.Namespace,
	})
	if err != nil {
		return fmt.Errorf("Failed to copy folder to pod: %s", err)
	}

	if c.Service {
		err = k8s.CreateOrUpdateService(ctx, clientset, k8s.CreateOrUpdateServiceParams{
			Name:              c.Name,
			Namespace:         c.Namespace,
			Port:              int32(c.Port),
			ContainerPort:     int32(c.ContainerPort),
			ReleaseIdentifier: releaseIdentifier,
		})
		if err != nil {
			return fmt.Errorf("Failed to create or update service: %s", err)
		}
	}

	if c.Ingress {
		err = k8s.CreateOrUpdateIngress(ctx, clientset, k8s.CreateOrUpdateIngressParams{
			Name:         c.Name,
			Namespace:    c.Namespace,
			IngressClass: &c.IngressClass,
			IngressHost:  c.IngressHost,
			Port:         int32(c.Port),
		})
		if err != nil {
			return fmt.Errorf("Failed to create or update ingress: %s", err)
		}
	}

	err = k8s.WaitForDeploymentToBeReady(ctx, clientset, k8s.WaitForDeploymentToBeReadyParams{
		Namespace:         c.Namespace,
		Name:              c.Name,
		ReleaseIdentifier: releaseIdentifier,
	})

	slog.Info("Deployment finished!")

	return nil
}
