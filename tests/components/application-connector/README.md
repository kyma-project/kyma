# Component tests for Application Connector

These are the component tests for Application Connector.

## Tests structure

The test are structured as a monorepo with tests for the following modules:
- Application Gateway

## Application Gateway tests

The tests are executed on a Kyma cluster where the tested Application Gateway is installed.

The environment consists of a Kubernetes Pod running the tests and a mock application simulating the remote endpoints for the tested Application Gateway.

![Application Gateway tests architecture](./assets/app-gateway-tests-architecture.svg)

The test cases are defined as services in the Application CR.

The whole test setup is deployed into the Kyma cluster with the Helm template executed in the Makefile by the command `test-gateway`.

Image versions and the external service name used during testing can be set up in the Helm chart values file `k8s/gateway-test/values.yaml`.

### Local build of test images

<!-- To build the test images locally, perform these steps: -->

1. Set up **DOCKER_TAG** and **DOCKER_PUSH_REPOSITORY** in `local_build.sh` for your target Docker registry settings.
2. Build the test images:
   ```bash
   ./local_build.sh
   ```

### Local execution

<!-- To run the tests locally, perform these steps: -->

1. Provision a local Kubernetes cluster with k3d:

   ```shell
   kyma provision k3d
   ```

2. Install the minimal set of components required to run Application Gateway for either Kyma OS or SKR:

   <div tabs name="Kyma flavor" group="minimal-kyma-installation">
      <details open>
      <summary label="OS">
      Kyma OS
      </summary>

   ```bash
   kyma deploy --components-file mini-kyma-os.yaml
   ```
      </details>
      <details>
      <summary label="SKR">
      SKR
      </summary>

   ```bash
   kyma deploy --components-file mini-kyma-skr.yaml 
   ```
      </details>
   </div>

4. Run the tests:

   ```bash
   make test-gateway
   ```
