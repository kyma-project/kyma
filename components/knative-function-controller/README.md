# Knative Function controller

## Overview

The Knative Function controller is a Kubernetes controller that enables Kyma to manage Function resources.

This Knative-based serverless implementation defines and handles the Function Custom Resource Definition with the help of Knative Build and Knative Serving.

## Prerequisites

- Knative Build (v0.6.0)
- Knative Serving (v0.6.1)
- Istio (istio-1.0.7)

## Installation


### Run locally
Follow these steps to run the Knative Function controller locally:

1. Create a ServiceAccount to enable knativeBuild Docker builds.

Currently you cannot run builds in every existing namespace of your cluster. ServiceAccounts with linked docker 
repository credentials have to be created for each namespace that will be used for the knative-function controller.

```bash
#set your namespace e.g. default
export NAMESPACE=<NAMESPACE>
#set your registry e.g. https://gcr.io
export REGISTRY=<YOURREGISTRY>
echo -n 'Username:' ; read username
echo -n 'Password:' ; read password
cat <<EOF | sed s/hihihih/hohohoho/
apiVersion: v1
kind: ServiceAccount
metadata:
    name: knative-function-controller-build
    labels:
        app: knative-function-controller
secrets:
    - name: knative-function-controller-docker-reg-credential
---
apiVersion: v1
kind: Secret
type: kubernetes.io/basic-auth
metadata:
    name: knative-function-controller-docker-reg-credential
    annotations:
        build.knative.dev/docker-0: ${REGISTRY}
data:
    username: $(echo $username | base64)
    password: $(echo $password | base64)
EOF
```

4. Install the CRD to a local Kubernetes cluster:

```bash
make install
```

5. Deploy the controller:
>**NOTE**: Run this command only when you want to deploy the controller locally. It is not necessary for production use.
```bash
eval $(minikube docker-env)
```

```bash
export DOCKER_TAG=<some tag e.g. latest>
export APP_NAME=knative-function-controller
export DOCKER_PUSH_REPOSITORY=<e.g. eu.gcr.io or index.docker.io>
export DOCKER_PUSH_DIRECTORY<e.g. pr or develop>
make install
make docker-build
make docker-push
make deploy
```
### Run on production
To use the controller on the production environment, uncomment the `manager_image_patch_remote_dev.yaml` line  in the `kustomization.yaml` file and follow the instructions for the local installation.
## Usage
### Test the controller
To test the controller, run:
```bash
make test
```

### Manage functions

Use the following examples to learn how to create and manage a function. 

Create a sample function:

```bash
kubectl apply -f config/samples/runtime_v1alpha1_function.yaml -n {NAMESPACE}
```

Search for a function:

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
