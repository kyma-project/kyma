# Knative Function controller

## Overview

The Knative Function controller is a Kubernetes Controller that enables Kyma to manage Function resources.

This Knative-based serverless implementation defines and handles the Function Custom Resource Definition with the help of Knative Build and Knative Serving.

## Prerequisites

- Knative Build (v0.6.0)
- Knative Serving (v0.6.1)
- Istio (istio-1.0.7)

## Installation


### Run locally

1. Modify the `config/config.yaml` file to include your base64-encoded `docker.io` credentials. 
2. Update the Docker registry value to your `docker.io` username.

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

This workflow needs to be used until kubernetes-sigs/kubebuilder#40 is fixed.

```bash
eval $(minikube docker-env)
make docker-build
make install
make deploy
```

#### Production Deployment

Uncomment `manager_image_patch_dev` in `kustomization.yaml`
Then run the following commands:

```bash
make install
APP_NAME=knative-function-controller
IMG={DOCKER_PUSH_REPOSITORY}/{DOCKER_PUSH_DIRECTORY}/{APP_NAME}
TAG={DOCKER_TAG}
make docker-push
make deploy
```

### Test
To test the controller, run:
```bash
make test
```

## Examples

Run the following examples to learn how to create and manage a function. 

Create a sample function:

```bash
kubectl apply -f config/samples/runtime_v1alpha1_function.yaml -n {NAMESPACE}
```

Search for a function

```bash
kubectl get functions -n {NAMESPACE}
```

```bash
kubectl get function -n {NAMESPACE}
```

```bash
kubectl get fcn -n {NAMESPACE}
```

Access the function:

```bash
export INGRESSGATEWAY=istio-ingressgateway
export IP_ADDRESS=$(kubectl get svc $INGRESSGATEWAY --namespace istio-system --output 'jsonpath={.status.loadBalancer.ingress[0].ip}')
curl -v -H "Host: $(kubectl get ksvc sample --output 'jsonpath={.status.domain}' -n {NAMESPACE}" http://$(minikube ip):$(kubectl get svc istio-ingressgateway --namespace istio-system --output 'jsonpath={.spec.ports[?(@.port==80)].nodePort}')
```
