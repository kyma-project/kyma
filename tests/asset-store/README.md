# Asset Store integration tests

## Overview

The project is a test suite for all Asset Store subcomponents, such as controllers and upload service.

## Prerequisites

Use the following tools to set up the project:

- [Go distribution](https://golang.org)
- [Docker](https://www.docker.com/)

## Usage

### Run a local version

To run the application against the local Kyma installation on Minikube without building the binary, expose Asset Upload Service from Kyma installation or run a local version, outside the Minikube cluster. Read the [Asset Upload Service Readme](../../components/asset-upload-service/README.md#usage) to learn, how to do that.

To run integration tests, use the following command:

```bash
APP_KUBECONFIG_PATH=/Users/$USER/.kube/config go test main_test.go
```

### Build a production version

To build the production Docker image, run this command:

```bash
docker build {image_name}:{image_tag}
```

The variables are:

- `{image_name}` that is the name of the output image. The default name is `asset-store-test`.
- `{image_tag}` that is the tag of the output image. The default tag is `latest`.

### Environmental variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| **APP_KUBECONFIG_PATH** | No |  | The path to the `kubeconfig` file, needed for running an application outside of the cluster |
| **APP_TEST_UPLOAD_SERVICE_URL** | No | `http://localhost:3000/v1/upload` | The URL of the Asset Upload Service |
| **APP_TEST_WAIT_TIMEOUT** | No | `3m` | The timeout of waiting for resources |
| **APP_TEST_NAMESPACE** | No | `test-asset-store` | The namespace name, created and deleted during integration tests |
| **APP_TEST_CLUSTER_BUCKET_NAME** | No | `test-cluster-bucket` | The ClusterBucket resource name |
| **APP_TEST_BUCKET_NAME** | No | `test-bucket` | The Bucket resource name |
| **APP_TEST_COMMON_ASSET_PREFIX** | No | `test` | The name prefix for Asset and ClusterAsset resources |

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:
```bash
dep ensure -vendor-only
```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, and checks the status of the vendored libraries. It also runs the static code analysis and ensures that the formatting of the code is correct.
