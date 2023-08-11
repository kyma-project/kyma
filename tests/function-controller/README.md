# Function Controller Integration Tests

## Overview

The project is a test scenario for the Function Controller. It creates a sample Function, exposes it using an APIRule, and sends `GET` requests to the Function to check if it is accessible from the cluster and outside of the cluster.

## Prerequisites

Use the following tools to set up the project:

- [Go v1.13](https://golang.org)
- [Docker](https://www.docker.com/)
- [Kyma CLI](https://github.com/kyma-project/cli)

## Usage

### Run a local version

To run integration tests, follow these instructions:

1. Install [Kyma](https://kyma-project.io/#/04-operation-guides/operations/02-install-kyma).
2. Build the test image directly on the Docker engine of the Minikube node without pushing it to a registry. Run:

   ```bash
   eval $(minikube docker-env)
   make build-image
   ```

   Alternatively, build the image and push it to a registry, such as Docker Hub.

3. Edit the TestDefinition CR and update its `.spec.template.spec.containers[0].image` field to `function-controller-test:latest` using this command:

   ```bash
   kubectl edit testdefinitions.testing.kyma-project.io -n kyma-system function-controller
   ```

4. Run the integration test. The command creates a test suite with a name in a form of `test-{ID}`. Run:

   ```bash
   kyma test run serverless
   ```

5. Get the test result using this command:

   ```bash
   kubectl logs -n kyma-system oct-tp-test-{ID}-function-controller-0 tests
   ```

### Build a production version

To build the production Docker image, run this command:

```bash
make build-image
```

### Environment variables

Use the following environment variables to configure the application:

| Name                                    | Required | Default                      | Description                                                                                                                   |
|-----------------------------------------| -------- |------------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| **APP_TEST_WAIT_TIMEOUT**               | No       | `5m`                         | The period of time for which the application waits for the resources to meet defined conditions                               |
| **APP_TEST_NAMESPACE_BASE_NAME**        | No       | `serverless`                 | The name of the Namespace used during integration tests                                                                       |
| **APP_TEST_FUNCTION_NAME**              | No       | `test-function`              | The name of the Function created and deleted during integration tests                                                         |
| **APP_TEST_APIRULE_NAME**               | No       | `test-apirule`               | The name of the APIRule created and deleted during integration tests                                                          |
| **APP_TEST_TRIGGER_NAME**               | No       | `test-trigger`               | The name of the Trigger created and deleted during integration tests                                                          |
| **APP_TEST_SERVICE_INSTANCE_NAME**      | No       | `test-service-instance`      | The name of the ServiceInstance created and deleted during integration tests                                                  |
| **APP_TEST_SERVICE_BINDING_NAME**       | No       | `test-service-binding`       | The name of the ServiceBinding created and deleted during integration tests                                                   |
| **APP_TEST_SERVICE_BINDING_USAGE_NAME** | No       | `test-service-binding-usage` | The name of the ServiceBindingUsage created and deleted during integration tests                                              |
| **APP_TEST_DOMAIN_NAME**                | No       | `test-function`              | The domain name used in the APIRule CR                                                                                        |
| **APP_TEST_INGRESS_HOST**               | No       | `kyma.local`                 | The Ingress host address                                                                                                      |
| **APP_TEST_DOMAIN_PORT**                | No       | `80`                         | The port of the Service exposed by the APIRule in a given domain                                                              |
| **APP_TEST_INSECURE_SKIP_VERIFY**       | No       | `true`                       | The flag that controls whether tests use verification of the server's certificate and the host name to reach the Function     |
| **APP_TEST_VERBOSE**                    | No       | `true`                       | The value that controls whether tests log resources that are subject to change                                                |
| **APP_TEST_MAX_POLLING_TIME**           | No       | `5m`                         | The maximum period of time in which the Function must reconfigure after an update                                             |
| **APP_TEST_KUBECTL_PROXY_ENABLED**      | No       | `false`                      | It enables running test locally with `kubectl proxy`. Run `kubectl proxy --proxy 8001` in the background and set the env to `true` |

## Development

### Install dependencies

This project uses `go modules` as a dependency manager. To install all required dependencies, use the following command:

```bash
go mod download
```
