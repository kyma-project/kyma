# Component tests for Application Connector

These are the component tests for Application Connector.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [Design and architecture](#design-and-architecture)
- [Building](#building)
- [Running](#running)
  - [Deploy Kyma cluster locally](#deploy-kyma-cluster-locally)
  - [Run the tests](#run-the-tests)
  - [Debugging](#debugging)

<!-- markdown-toc end -->

## Design and architecture

The tests consist of:
- [Application CRs](./resources/charts/gateway-test/templates/applications/) describing test cases
- [Test runners](./test/application-gateway/) with various check for subsets of cases, grouped by Application CRs
- [Mock application](./tools/external-api-mock-app/) which simulates remote endpoints

Additionally, following resources are created on the cluster:
- [Service Account](./resources/charts/gateway-test/templates/service-account.yml:2), used by tests to read Application CRs

The tests are executed as a Kubernetes Job on a Kyma cluster, 
where the tested Application Gateway is installed. 
The test job and mock application deployment are in the `test` namespace. 

![Application Gateway tests architecture](./assets/app-gateway-tests-architecture.svg)

## Building

To build Docker images of tests and mock application:

``` sh
./scripts/local-build.sh {DOCKER_TAG} {DOCKER_PUSH_REPOSITORY}
```
This will build the following images:
- `{DOCKER_PUSH_REPOSITORY}/gateway-test:{DOCKER_TAG}`
- `{DOCKER_PUSH_REPOSITORY}/mock-app:{DOCKER_TAG}`

## Running

### Deploy Kyma cluster locally

1. Provision a local Kubernetes cluster with k3d:
   ```sh
   kyma provision k3d
   ```

1. Install the minimal set of components required to run Application Gateway for either Kyma OS or SKR:

    <div tabs name="Kyma flavor" group="minimal-kyma-installation">
    <details open>
    <summary label="OS">
    Kyma OS
    </summary>

    ```sh
    kyma deploy --components-file ./resources/installation-config/mini-kyma-os.yaml
    ```

    </details>
    <details>
    <summary label="SKR">
    SKR
    </summary>

    ```bash
    kyma deploy --components-file ./resources/installation-config/mini-kyma-skr.yaml 
    ```

    </details>
    </div>

    >**TIP:** More on Kyma installation can be found in [official Kyma documentation](https://kyma-project.io/docs/kyma/latest/02-get-started/01-quick-install/#install-kyma)

### Run the tests

``` sh
make test-gateway
```

>**CAUTION:** This might override existing resources, if names of already existing resources collide with names used by tests

### Debugging
