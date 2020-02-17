# Rafter Integration Tests

## Overview

The project is a test scenario for all Rafter subcomponents, such as controllers and the Upload Service.

## Prerequisites

Use the following tools to set up the project:

- [Go](https://golang.org)
- [Docker](https://www.docker.com/)
- [Kyma CLI](https://github.com/kyma-project/cli)

## Usage

### Run a local version

To run integration tests, follow these instructions:

1. [Install](https://kyma-project.io/docs/master/root/kyma/#installation-install-kyma-locally) Kyma.
2. Build the test image directly on the Docker engine of the Minikube node without pushing it to a registry. Run:

   ```bash
   eval $(minikube docker-env)
   make build-image
   ```

   Alternatively, build the image and push it to a registry, such as Docker Hub.

3. Edit the TestDefinition CR and update its `.spec.template.spec.containers[0].image` field to `function-controller-test:latest` using this command:

   ```bash
   k edit testdefinitions.testing.kyma-project.io -n kyma-system function-controller
   ```

4. Run the integration test. The command creates a test suite with a name in a form of `test-{ID}`. Run:

   ```bash
   kyma test run function-controller
   ```

5. Get the test result using this command:

   ```bash
   k logs -n kyma-system oct-tp-test-{ID}-function-controller-0 tests
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
| **APP_TEST_WAIT_TIMEOUT**             | No       | `5m`                       | The period of time for which the application waits for the resources to meet defined conditions                                             |
| **APP_TEST_NAMESPACE**                | No       | `serverless`               | The name of the Namespace used during integration tests                                                                                     |
| **APP_TEST_FUNCTION_NAME**            | No       | `test-function`            | The name of the Function created and deleted during integration tests                                                                       |
| **APP_TEST_APIRULE_NAME**             | No       | `test-apirule`             | The name of the APIRule created and deleted during integration tests                                                                        |
| **APP_TEST_DOMAIN_NAME**              | No       | `test-function`            | The name of domain used in APIRule CR                                                                                                       |
| **APP_TEST_INGRESS_HOST**             | No       | `kyma.local`               | The Ingress host address                                                                                                                    |
| **APP_TEST_DOMAIN_PORT**              | No       | `80`                       | The port of domain                                                                                                                          |

Those can be supplied to [this](../../resources/function-controller/templates/tests/test.yaml) file before installing Kyma, or by editing TestDefinition CR with already installed Kyma using this command:

```bash
k edit testdefinitions.testing.kyma-project.io -n kyma-system function-controller
```

## Development

### Install dependencies

This project uses `go modules` as a dependency manager. To install all required dependencies, use the following command:

```bash
go mod download
```
