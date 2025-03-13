package command_test

import (
	"testing"
	"time"

	"github.com/lucasvmiguel/k8run/internal/command"
)

func TestDeploymentCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		command *command.DeploymentCommand
		wantErr bool
	}{
		{
			name: "valid command",
			command: &command.DeploymentCommand{
				Name:          "test-deployment",
				Image:         "test-image",
				CopyFolder:    "/test-folder",
				Replicas:      1,
				Timeout:       20 * time.Second,
				Service:       true,
				Port:          80,
				ContainerPort: 8080,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			command: &command.DeploymentCommand{
				Image:      "test-image",
				CopyFolder: "/test-folder",
				Replicas:   1,
				Timeout:    20 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing image",
			command: &command.DeploymentCommand{
				Name:       "test-deployment",
				CopyFolder: "/test-folder",
				Replicas:   1,
				Timeout:    20 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing copy folder",
			command: &command.DeploymentCommand{
				Name:     "test-deployment",
				Image:    "test-image",
				Replicas: 1,
				Timeout:  20 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid replicas",
			command: &command.DeploymentCommand{
				Name:       "test-deployment",
				Image:      "test-image",
				CopyFolder: "/test-folder",
				Replicas:   0,
				Timeout:    20 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			command: &command.DeploymentCommand{
				Name:       "test-deployment",
				Image:      "test-image",
				CopyFolder: "/test-folder",
				Replicas:   1,
				Timeout:    5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid service port",
			command: &command.DeploymentCommand{
				Name:          "test-deployment",
				Image:         "test-image",
				CopyFolder:    "/test-folder",
				Replicas:      1,
				Timeout:       20 * time.Second,
				Service:       true,
				Port:          -1,
				ContainerPort: 8080,
			},
			wantErr: true,
		},
		{
			name: "invalid ingress port",
			command: &command.DeploymentCommand{
				Name:         "test-deployment",
				Image:        "test-image",
				CopyFolder:   "/test-folder",
				Replicas:     1,
				Timeout:      20 * time.Second,
				Ingress:      true,
				Port:         -1,
				IngressHost:  "test-host",
				IngressClass: "test-class",
			},
			wantErr: true,
		},
		{
			name: "missing ingress host",
			command: &command.DeploymentCommand{
				Name:         "test-deployment",
				Image:        "test-image",
				CopyFolder:   "/test-folder",
				Replicas:     1,
				Timeout:      20 * time.Second,
				Ingress:      true,
				Port:         80,
				IngressClass: "test-class",
			},
			wantErr: true,
		},
		{
			name: "missing ingress class",
			command: &command.DeploymentCommand{
				Name:        "test-deployment",
				Image:       "test-image",
				CopyFolder:  "/test-folder",
				Replicas:    1,
				Timeout:     20 * time.Second,
				Ingress:     true,
				Port:        80,
				IngressHost: "test-host",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.command.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
