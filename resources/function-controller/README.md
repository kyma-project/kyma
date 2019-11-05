# Function Controller

## Overview

This project contains the chart for the Function Controller.

>**NOTE**: This feature is experimental.

## Prerequisites

- Knative Build (v0.6.0)
- Knative Serving (v0.8.1)
- Istio (v1.0.7)

## Details

### Install Helm chart

Follow the steps to install the chart:

1. Export the environment variables:

| Variable        | Description | Sample value | 
| --------------- | ----------- | --------|
| **FN_REGISTRY**   | The URL of the container registry Function images will be pushed to. Used for authentication.  | `https://gcr.io/` for GCR, `https://index.docker.io/v1/` for Docker Hub|
| **FN_REPOSITORY** | The name of the container repository Function images will be pushed to. | `gcr.io/my-project` for GCR, `my-user` for Docker Hub |

2. Install Knative Build:
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
3. Install the Function Controller charts:
    ```bash
    NAME=function-controller
    NAMESPACE=serverless-system
    
    FN_REGISTRY=https://index.docker.io/v1/
    FN_REPOSITORY=my-docker-user
    DOMAIN_NAME=kyma.local
    reg_username=<container registry username>
    reg_password=<container registry password>
    
    helm install function-controller \
                 --namespace="${NAMESPACE}" \
                 --name="${NAME}" \
                 --set secret.registryAddress="${FN_REGISTRY}" \
                 --set secret.registryUserName="${reg_username}" \
                 --set secret.registryPassword="${reg_password}" \
                 --set config.dockerRegistry="${FN_REPOSITORY}" \
                 --set global.ingress.domainName="${DOMAIN_NAME}" \
                 --tls
    ```
### Run the first function

Currently, there is no UI support for the new Function Controller. To use it, follow these steps:
1. Execute steps 4-6 from [these](https://github.com/kyma-project/kyma/blob/master/components/function-controller/README.md#installation) instructions.
2. Run your first function using [these](https://github.com/kyma-project/kyma/blob/master/components/function-controller/README.md#create-a-sample-hello-world-function) steps.

### Override default autoscaler configuration

The function controller uses [Knative Serving](https://github.com/kyma-project/kyma/tree/master/resources/knative-serving) under the hood. This means that the [Knative Pod Autoscaler (KPA)](https://knative.dev/docs/serving/configuring-the-autoscaler/) handles autoscaling by default. If you want to customize the settings, use [Helm overrides](https://kyma-project.io/docs/#configuration-helm-overrides-for-kyma-installation) to override the configuration defined in the `config-autoscaler` ConfigMap.