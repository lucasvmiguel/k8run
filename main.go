package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lucasvmiguel/k8run/internal/command"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "k8run",
		Usage:   "k8run is a CLI tool designed to quickly prototype Kubernetes deployments, services, and ingresses. It simplifies the process of setting up a working Kubernetes environment for development and testing.",
		Version: "0.1.0",
		Commands: []*cli.Command{
			{
				Name:      "destroy",
				Usage:     "Destroys a deployment with all its dependending resources",
				ArgsUsage: "<name>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "namespace",
						Usage:    "namespace to be used. eg: 'default'",
						Value:    "default",
						Required: false,
					},
					&cli.DurationFlag{
						Name:     "timeout",
						Usage:    "timeout for the deployment. eg: 30s",
						Required: false,
						Value:    time.Minute,
					},
					&cli.BoolFlag{
						Name:     "yes",
						Aliases:  []string{"y"},
						Usage:    "skips the confirmation",
						Required: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println()
					if !cmd.Bool("yes") && !confirm("Are you sure you want to proceed? (yes/no)") {
						fmt.Println("Operation aborted.")
						return nil
					}
					fmt.Println()

					c := command.NewDestroyCommand(command.NewDestroyCommandParams{
						Name:      cmd.Args().First(),
						Namespace: cmd.String("namespace"),
						Timeout:   cmd.Duration("timeout"),
					})

					if err := c.Validate(); err != nil {
						return err
					}

					return c.Run(ctx)
				},
			},
			{
				Name:      "deployment",
				Usage:     "Creates a deployment and dependending on the flags, a service and ingress",
				ArgsUsage: "<name>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "entrypoint",
						Usage:    "entrypoint of the container. eg: 'node index.js'",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "image",
						Usage:    "image to be used. eg: 'node:14'",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "copy",
						Usage:    "file or folder to be copied to the container. eg: '/Users/me/my_local_folder_to_copy'",
						Required: true,
					},
					&cli.BoolFlag{
						Name:     "service",
						Usage:    "if service will be created",
						Value:    false,
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "ingress",
						Usage:    "if ingress will be created",
						Value:    false,
						Required: false,
					},
					&cli.IntFlag{
						Name:     "container-port",
						Usage:    "port that the container is listening to",
						Required: false,
					},
					&cli.IntFlag{
						Name:     "port",
						Usage:    "port that the service will be listening to",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "ingress-class",
						Usage:    "ingress class to be used. eg: 'nginx'",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "ingress-host",
						Usage:    "ingress host to be used. eg: 'foo.myapp.com'",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "namespace",
						Usage:    "namespace to be used. eg: 'default'",
						Value:    "default",
						Required: false,
					},
					&cli.IntFlag{
						Name:     "replicas",
						Value:    1,
						Usage:    "number of replicas. eg: 3",
						Required: false,
					},
					&cli.DurationFlag{
						Name:     "timeout",
						Usage:    "timeout for the deployment. eg: 30s",
						Required: false,
						Value:    time.Minute,
					},
					&cli.BoolFlag{
						Name:     "yes",
						Aliases:  []string{"y"},
						Usage:    "skips the confirmation",
						Required: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println()
					if !cmd.Bool("yes") && !confirm("Are you sure you want to proceed? (yes/no)") {
						fmt.Println("Operation aborted.")
						return nil
					}
					fmt.Println()

					c := command.NewDeploymentCommand(command.NewDeploymentCommandParams{
						Name:       cmd.Args().First(),
						Namespace:  cmd.String("namespace"),
						Entrypoint: strings.Split(cmd.String("entrypoint"), " "),
						Timeout:    cmd.Duration("timeout"),
						// Deployment
						Replicas: int32(cmd.Int("replicas")),
						Copy:     cmd.String("copy"),
						Image:    cmd.String("image"),
						// Service
						Service:       cmd.Bool("service"),
						ContainerPort: cmd.Int("container-port"),
						Port:          cmd.Int("port"),
						// Ingress
						Ingress:      cmd.Bool("ingress"),
						IngressHost:  cmd.String("ingress-host"),
						IngressClass: cmd.String("ingress-class"),
					})

					if err := c.Validate(); err != nil {
						return err
					}

					return c.Run(ctx)
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

// confirm asks the user for confirmation (yes/no)
func confirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(message + " ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return false
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input == "yes" || input == "y" {
			return true
		} else if input == "no" || input == "n" {
			return false
		} else {
			fmt.Println("Please type 'yes' or 'no'.")
		}
	}
}
