# Function controller

## Overview

The Knative Function controller is a Kubernetes controller that enables Kyma to manage Function resources. It uses
Knative Build and Knative Serving under the hood.

## Prerequisites

- Knative Build (v0.6.0)
- Knative Serving (v0.6.1)
- Istio (v1.0.7)

## Installation for development workflow

### Preparation

A few preparation steps must be executed before proceeding with the rest of the instructions from this document.

#### Environment variables

The following environment variables must be exported to the current shell's environment.

| Variable        | Description |
| --------------- | ----------- |
| `IMG`           | Full image name the _Function Controller_ will be tagged with. (e.g. `gcr.io/my-project/function-controller` for GCR, `my-user/function-controller` for Docker Hub) |
| `FN_REGISTRY`   | URL of the container registry _Function_ images will be pushed to. Used for authentication. (e.g. `https://gcr.io/` for GCR, `https://index.docker.io/v1/` for Docker Hub) |
| `FN_REPOSITORY` | Name of the container repository _Function_ images will be pushed to. (e.g. `gcr.io/my-project` for GCR, `my-user` for Docker Hub) |

Example:

```bash
export IMG=my-docker-user/function-controller
export FN_REGISTRY=https://index.docker.io/v1/
export FN_REPOSITORY=my-docker-user
```

#### Controller namespace

Create the `serverless-system` Namespace. The controller will be deployed to it.

```bash
kubectl create namespace serverless-system
```

#### Container registry credentials

Before creating a Function, it is necessary to create the `registry-credentials` Secret, which contains credentials to
the Docker registry defined by the `FN_REGISTRY` environment variable. Knative Build uses this Secret to push the images
it builds for the deployed Functions.

```bash
reg_username=<container registry username>
reg_password=<container registry password>

cat <<EOF | kubectl -n serverless-system apply -f -
---
apiVersion: v1
kind: Secret
type: kubernetes.io/basic-auth
metadata:
  name: registry-credentials
  annotations:
    build.knative.dev/docker-0: ${FN_REGISTRY}
stringData:
  username: ${reg_username}
  password: ${reg_password}
EOF
```

#### Controller configuration

Create the following configuration for the controller. It contains a list of supported runtimes as well as the container
repository referenced by the `FN_REPOSITORY` environment variable, for which we created a Secret above.

```bash
cat <<EOF | kubectl -n serverless-system apply -f -
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config
data:
  serviceAccountName: function-controller
  dockerRegistry: ${FN_REPOSITORY}
  runtimes: |
    - ID: nodejs8
      dockerFileName: dockerfile-nodejs-8
    - ID: nodejs6
      dockerFileName: dockerfile-nodejs-6
EOF
```

### Deploy the controller

The following `make` targets build the Function Controller image, tag it to the value of the `IMG` environment variable
defined previously, and push it to the remote container registry.

```bash
make docker-build
make docker-push
```

After the image has been successfully pushed, one can deploy the controller to the `serverless-system` namespace created
in the previous section.

```bash
make deploy
```

## Usage

### Test the controller

The following `make` target runs tests against the deployed Function Controller.

```bash
make test
```

### Create a demo Function

> **Note**: In order to be able to create Functions outside of the `serverless-system` namespace (e.g. inside
> `default`), a few API objects must be copied from the `serverless-system` namespace:
> ```bash
> kubectl -n serverless-system get -o yaml \
>   secret/registry-credentials \
>   serviceaccount/function-controller \
>   configmap/fn-config \
>   configmap/dockerfile-nodejs-6 \
>   configmap/dockerfile-nodejs-8 \
> | sed -e '/namespace: /d'  -e '/-token-/d' \
> | kubectl apply -n my-namespace -f -
> ```

Deploy a sample _Hello World_ Function:

```bash
kubectl apply -f config/samples/serverless_v1alpha1_function.yaml
```

Ensure the Function was created:

```bash
kubectl get functions
```

Check the status of the build:

```bash
kubectl get builds.build.knative.dev
```

Check the status of the Knative Serving service:

```bash
kubectl get services.serving.knative.dev
```

Access the function:

```bash
# minikube
FN_DOMAIN="$(kubectl get ksvc demo --output 'jsonpath={.status.domain}')"
FN_PORT="$(kubectl get svc istio-ingressgateway -n istio-system --output 'jsonpath={.spec.ports[?(@.port==80)].nodePort}')"
curl -v -H "Host: ${FN_DOMAIN}" http://$(minikube ip):${FN_PORT}
```

```bash
# remote cluster (eg. GKE)
FN_DOMAIN="$(kubectl get ksvc demo --output 'jsonpath={.status.domain}')"
curl -kD- "https://${FN_DOMAIN}"
```
