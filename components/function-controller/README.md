# Function controller

## Overview

The Knative Function controller is a Kubernetes controller that enables Kyma to manage Function resources.

This Knative-based serverless implementation defines and handles the Function Custom Resource Definition with the help of Knative Build and Knative Serving.

## Prerequisites

- Knative Build (v0.6.0)
- Knative Serving (v0.6.1)
- Istio (v1.0.7)

## Installation for development workflow

Before installing the Function controller it is necessary to create the namespace, Service Account and Secret. The Service Account refers to the Secret containing the credentials to the Docker registry. Knative Build uses that Service Account to build and push the built images.
### Create Namespace
```bash
 NAMESPACE=<serverless-system>
kubectl create namespace serverless-system
```

### Create Service Account and Secret
```bash
#set your namespace e.g. default
NAMESPACE=<NAMESPACE>
#set your registry e.g. https://gcr.io or https://index.docker.io/v1/ (without the handle)
REGISTRY=<YOUR_REGISTRY>
#set your docker username and password
username=<docker username>
password=<docker password>
cat <<EOF | kubectl -n $NAMESPACE apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
    name: function-controller
    labels:
        app: function-controller
secrets:
    - name: function-controller-docker-reg-credential
---
apiVersion: v1
kind: Secret
type: kubernetes.io/basic-auth
metadata:
    name: function-controller-docker-reg-credential
    annotations:
        build.knative.dev/docker-0: ${REGISTRY}
data:
    username: $(echo $username | base64)
    password: $(echo $password | base64)
EOF
```

### Create the configuration file for the controller
```bash
DOCKER_REGISTRY_NAME=<YOUR_REGISTRY_HANDLE>
cat <<EOF | kubectl -n $NAMESPACE apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config
  labels:
    app: function-controller
data:
  serviceAccountName: function-controller
  dockerRegistry: $DOCKER_REGISTRY_NAME
  runtimes: |
    - ID: nodejs8
      dockerFileName: dockerfile-nodejs-8
    - ID: nodejs6
      dockerFileName: dockerfile-nodejs-6
EOF
```
### Run locally on Minikube
Follow these steps to run the Knative Function controller locally:
To use the controller locally in minikube, make sure the line `manager_image_patch_local_dev.yaml` is uncommented in the [kustomization file](config/default/kustomization.yaml).
### Deploy the controller
>**NOTE**: Run this command only when you want to deploy the controller locally. It is not necessary for production use.
```bash
eval $(minikube docker-env)
```

```bash
export DOCKER_TAG=<some tag e.g. latest>
export APP_NAME=function-controller
export DOCKER_PUSH_REPOSITORY=<e.g. eu.gcr.io or index.docker.io>
export DOCKER_PUSH_DIRECTORY=<e.g. pr or develop>
make install
make docker-build
make docker-push
make deploy
```
### Run on remote cluster (eg. GKE)
To use the controller in a remote cluster (eg. GKE), make sure the line `manager_image_patch_remote_dev.yaml` is uncommented in the [kustomization file](config/default/kustomization.yaml). Then follow the instructions [above](#deploy-the-controller).
## Usage
### Test the controller
To test the controller, run:
```bash
make test
```

### Manage functions

Use the following examples to learn how to create and manage a function. 

>**Note** If you are creating functions outside the `serverless-system` namespace then make sure following steps are followed:
```bash
NAMESPACE=foo
```
1. Create the `ServiceAccount` and `Secret` containing the credentials to the Docker registry in the namespace as [previously described](#create-service-account-and-secret).
2. Create the configmaps with docker templates for `node6` and `node8`
```bash
cat <<EOF | kubectl -n $NAMESPACE apply -f-
apiVersion: v1
kind: ConfigMap
metadata:
  name: dockerfile-nodejs-6
  labels:
    function: function-controller
data:
  Dockerfile: |-
    FROM kubeless/nodejs@sha256:5c3c21cf29231f25a0d7d2669c6f18c686894bf44e975fcbbbb420c6d045f7e7
    USER root
    RUN export KUBELESS_INSTALL_VOLUME='/kubeless' && \
        mkdir /kubeless && \
        cp /src/handler.js /kubeless && \
        cp /src/package.json /kubeless && \
        /kubeless-npm-install.sh
    USER 1000
EOF
```
```bash
cat <<EOF | kubectl -n $NAMESPACE apply -f-
apiVersion: v1
kind: ConfigMap
metadata:
  name: dockerfile-nodejs-8
  labels:
    function: function-controller
data:
  Dockerfile: |-
    FROM kubeless/nodejs@sha256:5c3c21cf29231f25a0d7d2669c6f18c686894bf44e975fcbbbb420c6d045f7e7
    USER root
    RUN export KUBELESS_INSTALL_VOLUME='/kubeless' && \
        mkdir /kubeless && \
        cp /src/handler.js /kubeless && \
        cp /src/package.json /kubeless && \
        /kubeless-npm-install.sh
    USER 1000
EOF
```
```bash
kubectl apply -f config/samples/serverless_v1alpha1_function.yaml -n {NAMESPACE}
```

Search for a function:
```bash
kubectl get functions -n ${NAMESPACE}
```

Check the status of the build
```bash
kubectl get builds.build.knative.dev -n ${NAMESPACE}
```

Check the status of the serving service
```bash
kubectl get services.serving.knative.dev -n ${NAMESPACE}
```

Access the function:

On minikube
```bash
export INGRESSGATEWAY=istio-ingressgateway
export IP_ADDRESS=$(kubectl get svc $INGRESSGATEWAY --namespace istio-system --output 'jsonpath={.status.loadBalancer.ingress[0].ip}')
curl -v -H "Host: $(kubectl get ksvc sample --output 'jsonpath={.status.domain}' -n ${NAMESPACE}" http://$(minikube ip):$(kubectl get svc istio-ingressgateway --namespace istio-system --output 'jsonpath={.spec.ports[?(@.port==80)].nodePort}')
```

on Remote Cluster (eg. GKE)
```bash
curl -v "https://$(kubectl get ksvc -n ${NAMESPACE} --output 'jsonpath={.items[0].status.domain}')"
```
