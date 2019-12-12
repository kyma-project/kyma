# Asset Store Integration Tests

## Overview

The project is a test scenario for all Rafter subcomponents, such as controllers and the Asset Upload Service.

## Prerequisites

Use the following tools to set up the project:

- [Go](https://golang.org)
- [Docker](https://www.docker.com/)

## Usage

### Run a local version

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

- `{image_name}` that is the name of the output image.
- `{image_tag}` that is the tag of the output image. The default tag is `latest`.

### Environmental variables

Use the following environment variables to configure the application:

| Name                                  | Required | Default                    | Description                                                                                     |
| ------------------------------------- | -------- | -------------------------- | ----------------------------------------------------------------------------------------------- |
| **APP_KUBECONFIG_PATH**               | No       | None                       | The path to the `kubeconfig` file, needed for running an application outside of the cluster     |
| **APP_TEST_WAIT_TIMEOUT**             | No       | `3m`                       | The period of time for which the application waits for the resources to meet defined conditions |
| **APP_TEST_NAMESPACE**                | No       | `test-asset-store`         | The name of the Namespace created and deleted during integration tests                          |
| **APP_TEST_CLUSTER_BUCKET_NAME**      | No       | `test-cluster-bucket`      | The ClusterBucket resource name                                                                 |
| **APP_TEST_BUCKET_NAME**              | No       | `test-bucket`              | The Bucket resource name                                                                        |
| **APP_TEST_ASSET_GROUP_NAME**         | No       | `test-asset-group`         | The AssetGroup resource name                                                                    |
| **APP_TEST_CLUSTER_ASSET_GROUP_NAME** | No       | `test-cluster-asset-group` | The ClusterAssetGroup resource name                                                             |
| **APP_TEST_COMMON_ASSET_PREFIX**      | No       | `test`                     | The name of the prefix for the Asset and ClusterAsset resources                                 |
| **APP_TEST_MOCKICE_NAME**             | No       | `rafter-test-svc`          | The name of the pod, service and configmap used by test service                                 |

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:

```bash
dep ensure -vendor-only
```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, and checks the status of the vendored libraries. It also runs the static code analysis and ensures that the formatting of the code is correct.
