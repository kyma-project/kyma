# knative-function-controller

## Overview

This project contains the chart for the Function Controller.

>**NOTE**: This feature is experimental.

## Prerequisites

- Knative Build (v0.6.0)
- Knative Serving (v0.6.1)
- Istio (istio-1.0.7)

## Installation

### Install Helm chart

Run the following script to install the chart:

```bash
export NAME=knative-function-controller
export NAMESPACE=kyma-system
helm install knative-function-controller \
             --namespace="${NAMESPACE}" \
             --name="${NAME}" \
             --tls
```


### Create a ServiceAccount to enable knativeBuild Docker builds


To run a build in a Namespace, create ServiceAccounts with linked Docker repository credentials for each Namespace that will be used for the knative-function controller.
```bash
#set your namespace e.g. default
export NAMESPACE=<NAMESPACE>
#set your registry e.g. https://gcr.io
export REGISTRY=<YOURREGISTRY>
echo -n 'Username:' ; read username
echo -n 'Password:' ; read password
cat <<EOF | kubectl apply -n ${NAMESPACE} -f -
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

## Running the first function

Currently there is no UI support for the new knative-function-controller.
Run your first function in the following way:

```bash
export NAMESPACE=<NAMESPACE>
cat <<EOF | kubectl apply -n ${NAMESPACE} -f -
apiVersion: runtime.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: sample
  labels:
    foo: bar
spec:
  function: |
    module.exports = {
        main: function(event, context) {
          return 'Hello World'
        }
      }
  functionContentType: "plaintext"
  size: "L"
  runtime: "nodejs8"
EOF
``` 

To get the URL of your new service, run:
```bash
kubectl get ksvc -n $NAMESPACE
```
