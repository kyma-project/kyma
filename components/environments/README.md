## Overview

This controller injects limit ranges, resource quotas, and default roles into each Namespace that you create.

Developers create default roles from a roles template that they first define in a Namespace inside the Kubernetes cluster.
At the time of cluster provisioning, developers might define the roles in the `3-bootstrap-roles.yaml` file. The controller looks for roles labeled as `env=true` at the creation of the new Namespace. Next, the controller copies the roles to the new Namespace.

Limit range configuration is required. These environment variables provide the configuration:
* `APP_LIMIT_RANGE_MEMORY_DEFAULT_REQUEST`
* `APP_LIMIT_RANGE_MEMORY_DEFAULT`
* `APP_LIMIT_RANGE_MEMORY_MAX`

Each Kubernetes environment has a `ResourceQuota` configured with standard quantities (such as 1G1 or 256Mi) that are set in the following required configuration values:
* `APP_RESOURCE_QUOTA_REQUESTS_MEMORY`
* `APP_RESOURCE_QUOTA_LIMITS_MEMORY`

## Prerequisites

 - A working Golang installation
 - Minikube
 - kubectl
 - Docker

### Build to run on Kubernetes

Use the following commands to prepare to run on Kubernetes. Run them in the following order:

 - `dep ensure`
 - `export GOOS=linux`
 - `go build -o environments cmd/controller/main.go`

### Build a Docker image

Make sure that the [build](#build-to-run-on-kubernetes) step is complete. Run the following commands:

- `cp ./environments deploy/controller/environments`
- `docker build -t environments:{your_tag} deploy/controller`

Make sure the image is built:

- `docker images | grep environments`

### Run the image locally inside Kyma

This section describes how to run Kyma with an updated environments image. The procedure is useful in case the component has been modified and needs to be tested.

Read the main [Kyma project README.md](../../README.md). By default, the system runs the environments image specified in the [4-deployment.yaml](../../resources/core/charts/environments/templates/4-deployment.yaml) file. You can provide your own image by following one of the procedures.

#### Docker registry

If you have access to an external Docker registry, build your Docker image, push it to the registry and modify the [4-deployment.yaml](../../resources/core/charts/environments/templates/4-deployment.yaml) file by swapping the evironments image. Follow the [instructions](../../docs/kyma/docs/031-gs-local-installation.md) to run Kyma as usual.

#### Minikube built in Docker daemon

In case you have no access to a Docker registry, use Minikube’s built in Docker daemon that keeps images for running containers:

1. Modify the [4-deployment.yaml](../../resources/core/charts/environments/templates/4-deployment.yaml) file by swapping the evironments image.
```
image: environments:my_tag
```

2. [Start Kyma installation as usual](../../docs/kyma/docs/031-gs-local-installation.md).

3. Run the following command to set up the Docker environment variables so a Docker client can communicate with the Minikube Docker daemon:
```
eval $(minikube docker-env)
```

4. [Build your Docker image](#build-a-docker-image) with the tag you specified in the first step. Wait for the installation to complete.
