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

Before you run the acceptance tests, export required environmental variables with the following command:

```bash
export USERNAME=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.email}" | base64 -D)
export PASSWORD=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 -D)
export KUBECONFIG=/Users/${USER}/.kube/config
```

Run acceptance tests using the following command:

- against the UI API Layer deployed on the local cluster:
  
  ```bash
  go test ./... -tags=acceptance
  ```

- against standalone UI API Layer deployed on the local host:
  
  ```bash
  GRAPHQL_ENDPOINT=http://localhost:3000/graphql go test ./... -tags=acceptance
  ```

- against the UI API Layer deployed on the cluster with custom domain:
  
  ```bash
  DOMAIN=nightly.kyma.cx go test ./... -tags=acceptance
  ```
  
To run the tests against the UI API Layer with module pluggability turned on, add the following environmental variable:
  
```bash
MODULE_PLUGGABILITY=true
```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, checks the status of the vendored libraries, runs the static code analysis, and ensures that the formatting of the code is correct.
