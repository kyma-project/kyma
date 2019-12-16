# Rafter Integration Tests

## Overview

The project is a test scenario for all Rafter subcomponents, such as controllers and the Asset Upload Service.

## Prerequisites

Use the following tools to set up the project:

- [Go](https://golang.org)
- [Docker](https://www.docker.com/)
- [Kyma CLI](https://github.com/kyma-project/cli)

## Usage

### Run a local version

To run integration tests, follow those instructions:

1. [Install](https://kyma-project.io/docs/master/root/kyma/#installation-install-kyma-locally) Kyma
2. Build test image directly on the Docker engine of the Minikube node without pushing it to a registry. To build the image on Minikube:

   ```bash
   eval $(minikube docker-env)
   make build-image
   ```

   Alternatively, build the image and push it to some registry, for example Docker Hub.

3. Edit TestDefinition CR and update its `.spec.template.spec.containers[0].image` field to `rafter-test:latest` using this command:

   ```bash
   k edit testdefinitions.testing.kyma-project.io -n kyma-system rafter
   ```

4. Run intergration test using this command:

   ```bash
   kyma test run rafter
   ```

   This creates test suite with name in a form of `test-{ID}`.

5. Get the test result using this command:

   ```bash
   k logs -n kyma-system oct-tp-test-{ID}-rafter-0 tests
   ```

### Build a production version

To build the production Docker image, run this command:

```bash
make build-image
```

### Environmental variables

Use the following environment variables to configure the application:

| Name                                  | Required | Default                    | Description                                                                                                                                 |
| ------------------------------------- | -------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| **APP_KUBECONFIG_PATH**               | No       | None                       | The path to the `kubeconfig` file, needed for running an application outside of the cluster. If not supplied in-cluster config will be used |
| **APP_TEST_WAIT_TIMEOUT**             | No       | `3m`                       | The period of time for which the application waits for the resources to meet defined conditions                                             |
| **APP_TEST_NAMESPACE**                | No       | `rafter-test`              | The name of the Namespace created and deleted during integration tests                                                                      |
| **APP_TEST_CLUSTER_BUCKET_NAME**      | No       | `test-cluster-bucket`      | The ClusterBucket resource name                                                                                                             |
| **APP_TEST_BUCKET_NAME**              | No       | `test-bucket`              | The Bucket resource name                                                                                                                    |
| **APP_TEST_ASSET_GROUP_NAME**         | No       | `test-asset-group`         | The AssetGroup resource name                                                                                                                |
| **APP_TEST_CLUSTER_ASSET_GROUP_NAME** | No       | `test-cluster-asset-group` | The ClusterAssetGroup resource name                                                                                                         |
| **APP_TEST_COMMON_ASSET_PREFIX**      | No       | `test`                     | The name of the prefix for the Asset and ClusterAsset resources                                                                             |
| **APP_TEST_MOCKICE_NAME**             | No       | `rafter-test-svc`          | The name of the pod, service and configmap used by test service                                                                             |

Those can be supplied to [this](../../resources/rafter/templates/tests/test.yaml) file before installing Kyma, or by editing TestDefinition CR with already installed Kyma using this command:

```bash
k edit testdefinitions.testing.kyma-project.io -n kyma-system rafter
```

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:

```bash
dep ensure -vendor-only
```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, and checks the status of the vendored libraries. It also runs the static code analysis and ensures that the formatting of the code is correct.
