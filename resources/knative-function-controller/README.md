# knative-function-controller

## Overview

This project contains the chart for the Function Controller.

## Prerequisites

- Istio
- Knative Serving
- Knative Build

## Installation

Run the following script to install the chart:

```bash
export NAME=knative-function-controller
export NAMESPACE=kyma-system
export REGISTRY_ADDRESS=<e.g. https://eu.gcr.io>
export REGISTRY_USER_NAME=<eu.gcr.io username goes here. Not Email>
export REGISTRY_PASSWORD=<password of the registry. e.g. echo -n $PASSWORD | base64>
helm install knative-function-controller --set secret.registryAddress="${REGISTRY_ADDRESS}" \
             --set secret.registryUserName="${REGISTRY_USER_NAME}" \
             --set secret.registryPassword="${REGISTRY_PASSWORD}" \
             --namespace="${NAMESPACE}" \
             --name="${NAME}"
```
