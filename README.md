# k8run

**k8run** is a CLI tool designed to quickly prototype Kubernetes deployments, services, and ingresses. It simplifies the process of setting up a working Kubernetes environment for development and testing.

> ⚠️ **Warning**: Although this tool may work fine in a production environment, it’s not recommended since it was specifically designed for testing and prototyping applications.

## Why?

Containerization offers many advantages in a production environment, but sometimes you just want to quickly test or prototype something without the hassle of building and publishing a Docker image. This tool is designed to make deploying to Kubernetes quick and easy.

## Requirements

* [kubectl](https://kubernetes.io/docs/reference/kubectl/)

## Features
- Create Kubernetes **Deployments** easily.
- Automatically create **Services** to expose your applications.
- Configure **Ingresses** with custom hosts and classes.
- Specify container images, ports, and entry points.
- Copy local folders into the container for easy prototyping.


## Installation

```bash
bash <(curl -sSL https://raw.githubusercontent.com/lucasvmiguel/k8run/main/install.sh)
```

if you have Golang installed, you can also install by running:
```bash
go install github.com/lucasvmiguel/k8run@latest
```

You can also download the binary [here](https://github.com/lucasvmiguel/k8run/releases)

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

## Roadmap

* Add video of how to use the tool
* Remove dependency of kubectl
* Add integration tests
* Add destroy command
* Add job command
* Add cronjob command

## Release

1. Change the version on [Makefile](Makefile)
2. Run `make release`

## License

This project is licensed under the [MIT License](LICENSE)