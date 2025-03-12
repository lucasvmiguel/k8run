# k8run

**k8run** is a CLI tool designed to quickly prototype Kubernetes deployments, services, and ingresses. It simplifies the process of setting up a working Kubernetes environment for development and testing.


## Requirements

* [kubectl](https://kubernetes.io/docs/reference/kubectl/)

## Features
- Create Kubernetes **Deployments** easily.
- Automatically create **Services** to expose your applications.
- Configure **Ingresses** with custom hosts and classes.
- Specify container images, ports, and entry points.
- Copy local folders into the container for easy prototyping.


## Installation
Clone the repository and build the CLI:

```bash
git clone https://github.com/lucasvmiguel/k8run.git
cd k8run
go build -o k8run
```

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



## How It Works
	1.	Generates Kubernetes YAML manifests for Deployment, Service, and Ingress.
	2.	Applies the manifests using kubectl.
	3.	Copies files into the container using kubectl cp (if specified).

## License

This project is licensed under the [MIT License](LICENSE)