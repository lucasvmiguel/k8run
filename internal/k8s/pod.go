package k8s

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"log/slog"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WaitForRunningInitContainerParams represents the parameters to wait for a running init container.
type WaitForRunningInitContainerParams struct {
	Namespace         string
	Name              string
	InitContainerName string
	ReleaseIdentifier string
}

// WaitForRunningInitContainer waits for the init container to be running.
func WaitForRunningInitContainer(ctx context.Context, clientset kubernetes.Interface, params WaitForRunningInitContainerParams) (*corev1.Pod, error) {
	sleep := 5 * time.Second
	podsClient := clientset.CoreV1().Pods(params.Namespace)

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting for init container to be running")
		default:
			slog.With("name", params.Name, "namespace", params.Namespace).Info("Waiting for init container to be running...")

			pods, err := podsClient.List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", LabelNameReleaseIdentifier, params.ReleaseIdentifier)})
			if err != nil {
				return nil, fmt.Errorf("failed to list pods: %w", err)
			}

			if len(pods.Items) == 0 {
				time.Sleep(sleep)
				continue
			}

			for _, pod := range pods.Items {
				if params.ReleaseIdentifier == pod.Labels[LabelNameReleaseIdentifier] {
					for _, containerStatus := range pod.Status.InitContainerStatuses {
						if containerStatus.Name == params.InitContainerName && containerStatus.State.Running != nil {
							slog.With("pod", pod.Name).Info("Init container is running.")
							return &pod, nil
						}
					}
				}
			}

			time.Sleep(sleep)
		}
	}
}

// CopyToPodParams represents the parameters to copy a folder to a pod.
type CopyToPodParams struct {
	LocalPath         string
	PodName           string
	ContainerPath     string
	InitContainerName string
	Namespace         string
}

// CopyToPod copies a file or folder to a pod.
func CopyToPod(params CopyToPodParams) error {
	slog.With("podName", params.PodName, "namespace", params.Namespace).Info("Copying to pod...")

	cmd := exec.Command("kubectl", "cp", params.LocalPath, fmt.Sprintf("%s:%s", params.PodName, params.ContainerPath), "-c", params.InitContainerName, "-n", params.Namespace)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error copying folder: %s, output: %s", err, string(output))
	}

	return nil
}
