# UI API Layer Acceptance tests

## Overview

This project includes acceptance tests for a UI API Layer project.

## Prerequisites

Use the following tools to set up the project:

* [Go distribution](https://golang.org)
* [Docker](https://www.docker.com/)

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:
```bash
dep ensure -vendor-only
```

### Run tests outside the cluster

To run acceptance tests outside the cluster, use this command to enable port forwarding for the Helm client:
```bash
kubectl port-forward $(kubectl get po -n kube-system | grep tiller |  awk '{print $1}') 44134:44134 -n kube-system
```

Run acceptance tests using the following command:

- against the UI API Layer deployed on the local cluster:
  
  ```bash
  KUBE_CONFIG=/Users/${USER}/.kube/config go test ./... -tags=acceptance
  ```

- against standalone UI API Layer deployed on the local host:
  
  ```bash
  KUBE_CONFIG=/Users/${USER}/.kube/config IS_LOCAL_CLUSTER=false GRAPHQL_ENDPOINT=http://localhost:3000/graphql go test ./... -tags=acceptance
  ```

- against the UI API Layer deployed on the cluster with custom domain:
  
  ```bash
  KUBE_CONFIG=/Users/${USER}/.kube/config IS_LOCAL_CLUSTER=false DOMAIN=nightly.kyma.cx go test ./... -tags=acceptance
  ```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, checks the status of the vendored libraries, runs the static code analysis, and ensures that the formatting of the code is correct.
