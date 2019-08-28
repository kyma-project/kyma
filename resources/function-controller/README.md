# function-controller

## Overview

This project contains the chart for the Function Controller.

>**NOTE**: This feature is experimental.

## Prerequisites

- Knative Build (v0.6.0)
- Knative Serving (v0.6.1)
- Istio (v1.0.7)

## Installation

### Install Helm chart

Environment variables:

| Variable        | Description |
| --------------- | ----------- |
| `FN_REGISTRY`   | URL of the container registry _Function_ images will be pushed to. Used for authentication. (e.g. `https://gcr.io/` for GCR, `https://index.docker.io/v1/` for Docker Hub) |
| `FN_REPOSITORY` | Name of the container repository _Function_ images will be pushed to. (e.g. `gcr.io/my-project` for GCR, `my-user` for Docker Hub) |

1. Install knative-build
    ```bash
    helm install knative-build-init \
                 --namespace="knative-build" \
                 --name="knative-build-init" \
                 --tls
    
    helm install knative-build \
                 --namespace="knative-build" \
                 --name="knative-build" \
                 --tls
    ```
2. Install the function controller charts
    ```bash
    NAME=function-controller
    NAMESPACE=serverless-system
    
    FN_REGISTRY=https://index.docker.io/v1/
    FN_REPOSITORY=my-docker-user
    reg_username=<container registry username>
    reg_password=<container registry password>
    
    helm install function-controller \
                 --namespace="${NAMESPACE}" \
                 --name="${NAME}" \
                 --set secret.registryAddress="${FN_REPOSITORY}" \
                 --set secret.registryUserName="${reg_username}" \
                 --set secret.registryPassword="${reg_password}" \
                 --set config.dockerRegistry="${FN_REPOSITORY}" \
                 --tls
    ```
## Running the first function

Currently there is no UI support for the new function-controller.
Run your first function as mentioned [here](../../components/function-controller/README.md#create-a-sample-hello-world-function)