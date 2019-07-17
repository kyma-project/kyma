# Knative Function Controller

Modify the project title and insert the name of your project. Use Heading 1 (H1).

## Overview

The Knative Function Controller is a Kubernetes Controller that enable Kyma to manage Function resources.

It defines and handles a function Custom Resource Definition with the help of Knative Build and Knative Serving. Basically it is the serverless implementation based on knative.

## Prerequisites

- Knative Build (v0.6.0)
- Knative Serving (v0.6.1)
- Istio (istio-1.0.7)

## Installation

### Local Deployment

The Manager running locally.

modify config/config.yaml to include your docker.io credentials (base64 encoded) and update the docker registry value to your docker.io username

Apply the configuration

```bash
kubectl apply -f config/config.yaml
```

Install the CRD to a local Kubernetes cluster:

```bash
make install
```

Run the controller on your machine:

```bash
make run
```

#### Manager running inside k8s cluster

This workflow needs to be used until [Come up with webhook developer workflow to test it locally #400](https://github.com/kubernetes-sigs/kubebuilder/issues/400) is fixed.

```bash
eval $(minikube docker-env)
make docker-build
make install
make deploy
```

#### Prod Deployment

Uncomment `manager_image_patch_dev` in `kustomization.yaml`
Then run the following commands:

```bash
make install
APP_NAME = knative-function-controller
IMG = {DOCKER_PUSH_REPOSITORY}/{DOCKER_PUSH_DIRECTORY}/{APP_NAME}
TAG = {DOCKER_TAG}
make docker-push
make deploy
```

### Test

```bash
make test
```

#### Examples