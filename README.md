# k8run

**k8run** is a CLI tool designed to quickly prototype Kubernetes deployments, services, and ingresses. It simplifies the process of setting up a working Kubernetes environment for development and testing.

> ⚠️ **Warning**: Although this tool may work fine in a production environment, it’s not recommended since it was specifically designed for testing and prototyping applications.

## Why?

Containerization offers many advantages in a production environment, but sometimes you just want to quickly test or prototype something without the hassle of building and publishing a Docker image. This tool is designed to make deploying to Kubernetes quick and easy.

## Requirements

* [kubectl](https://kubernetes.io/docs/reference/kubectl/)

## Installation

```bash
bash <(curl -sSL https://raw.githubusercontent.com/lucasvmiguel/k8run/main/install.sh)
```

Other ways:

* [Install binaries](https://github.com/lucasvmiguel/k8run/releases)
* `go install github.com/lucasvmiguel/k8run@latest` (in case you have Golang installed)

## Features
- Create Kubernetes **Deployments** easily.
- Automatically create **Services** to expose your applications.
- Configure **Ingresses** with custom hosts and classes.
- Specify container images, ports, and entry points.
- Copy local folders into the container for easy prototyping.

## How it works

**k8run** simplifies Kubernetes deployments by using an init container to handle the setup process. When a deployment starts, the init container continuously monitors the contents of a specified folder. Simultaneously, k8run copies the folder specified by the `--copy-folder` label into the init container. Once the init container detects that the folder is no longer empty, it exits, signaling that the setup is complete. At this point, the main container (defined by the `--image` label) starts executing with the specified entry point (set via the `--entrypoint` label).

## Usage

### Create a deployment (optionally creates a service and ingress)

Usage:

```bash
NAME:
   k8run deployment - Creates a deployment and dependending on the flags, a service and ingress

USAGE:
   k8run deployment [command [command options]] <name>

OPTIONS:
   --entrypoint value      entrypoint of the container. eg: 'node index.js'
   --image value           image to be used. eg: 'node:14'
   --copy-folder value     folder to be copied to the container. eg: '/Users/me/my_local_folder_to_copy'
   --service               if service will be created (default: false)
   --ingress               if ingress will be created (default: false)
   --container-port value  port that the container is listening to (default: 0)
   --port value            port that the service will be listening to (default: 0)
   --ingress-class value   ingress class to be used. eg: 'nginx'
   --ingress-host value    ingress host to be used. eg: 'foo.myapp.com'
   --namespace value       namespace to be used. eg: 'default' (default: "default")
   --replicas value        number of replicas. eg: 3 (default: 1)
   --timeout value         timeout for the deployment. eg: 30s (default: 30s)
   --yes, -y               skips the confirmation (default: false)
   --help, -h              show help
```

Example:

```bash
k8run deployment foobar \
  --service \
  --container-port 3000 \
  --port 8080 \
  --ingress \
  --ingress-class traefik \
  --ingress-host foobar.myproject.me \
  --image node \
  --namespace default \
  --entrypoint "node index.js" \
  --copy-folder /Users/myuser/projects/foobar
```

### Destroy a deployment (also destroys all resources associated with it)

Usage:

```bash
NAME:
   k8run destroy - Destroys a deployment with all its dependending resources

USAGE:
   k8run destroy [command [command options]] <name>

OPTIONS:
   --namespace value  namespace to be used. eg: 'default' (default: "default")
   --timeout value    timeout for the deployment. eg: 30s (default: 1m0s)
   --yes, -y          skips the confirmation (default: false)
   --help, -h         show help
```

Example:

```bash
k8run deployment foobar --namespace default
```


## Roadmap

* Add video of how to use the tool
* Remove dependency of kubectl
* Add integration tests
* Add job command
* Add cronjob command

## Release

1. Change the version on [Makefile](Makefile)
2. Run `make release`

## License

This project is licensed under the [MIT License](LICENSE)