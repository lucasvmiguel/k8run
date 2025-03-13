package command_test

import (
	"testing"
	"time"

	"github.com/lucasvmiguel/k8run/internal/command"
)

func TestDestroyCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		command *command.DestroyCommand
		wantErr bool
	}{
		{
			name: "valid command",
			command: &command.DestroyCommand{
				Name:      "test",
				Namespace: "default",
				Timeout:   15 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			command: &command.DestroyCommand{
				Name:      "",
				Namespace: "default",
				Timeout:   15 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "timeout too short",
			command: &command.DestroyCommand{
				Name:      "test",
				Namespace: "default",
				Timeout:   5 * time.Second,
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
