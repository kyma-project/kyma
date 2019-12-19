# Console Backend Service Test

## Overview

This project includes acceptance tests for a Console Backend Service project.

## Prerequisites

Use the following tools to set up the project:

- [Go](https://golang.org)
- [Docker](https://www.docker.com/)

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:

```bash
dep ensure -vendor-only
```

### Run tests outside the cluster

Before you run the acceptance tests, export required environment variables with the following command:

```bash
export KUBECONFIG=/Users/${USER}/.kube/config && export ADMIN_EMAIL=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode) && export ADMIN_PASSWORD=$(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode) && export READ_ONLY_USER_PASSWORD=$(kubectl get secret test-read-only-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode) && export READ_ONLY_USER_EMAIL=$(kubectl get secret test-read-only-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode) && export NO_RIGHTS_USER_PASSWORD=$(kubectl get secret test-no-rights-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode) && export NO_RIGHTS_USER_EMAIL=$(kubectl get secret test-no-rights-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode) && export TEST_TESTING_ADDONS_URL="https://github.com/kyma-project/addons/releases/download/0.8.0/index-testing.yaml"
```

Run acceptance tests using the following command:

- against the Console Backend Service deployed on the local cluster:

  ```bash
  go test ./... -tags=acceptance -v -count=1 -p=1
  ```

- against standalone Console Backend Service deployed on the local host:

  ```bash
  GRAPHQL_ENDPOINT=http://localhost:3000/graphql go test ./... -tags=acceptance -v -count=1 -p=1
  ```

- against the Console Backend Service deployed on the cluster with custom domain:

  ```bash
  DOMAIN=nightly.kyma.cx go test ./... -tags=acceptance -v -count=1 -p=1
  ```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, checks the status of the vendored libraries, runs the static code analysis, and ensures that the formatting of the code is correct.
