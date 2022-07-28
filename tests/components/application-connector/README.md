# Component tests for Application Connector

These are the component tests for Application Connector.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [Component tests for Application Connector](#component-tests-for-application-connector)
- [Design and architecture](#design-and-architecture)
- [Building](#building)
- [Running](#running)
    - [Deploy Kyma cluster locally](#deploy-kyma-cluster-locally)
    - [Run the tests](#run-the-tests)
- [Debugging](#debugging)
    - [Running locally](#running-locally)
    - [Running without cleanup](#running-without-cleanup)

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
The test job and mock application deployment are in the `test` Namespace. 

![Application Gateway tests architecture](./assets/app-gateway-tests-architecture.svg)

## Building

To build **and push** Docker images of tests and mock application:

``` sh
./scripts/local-build.sh {DOCKER_TAG} {DOCKER_PUSH_REPOSITORY}
```
This will build the following images:
- `{DOCKER_PUSH_REPOSITORY}/gateway-test:{DOCKER_TAG}`
- `{DOCKER_PUSH_REPOSITORY}/mock-app:{DOCKER_TAG}`

## Running

Tests can be run on any Kyma cluster with Application Gateway.

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

By default tests clean up after themselves, removing all created resources
and `test` Namespace.

> **CAUTION:** This might override or remove existing resources, 
> if names of already existing resources collide with names used by tests

## Debugging

### Running locally

The Test Job must run on a cluster, because of the way it accesses Application CRs.
Application Gateway and Mock Application can both be run locally.

To run Mock Application locally:
1. Change all `targetUrl` in [Application CRs](./resources/charts/gateway-test/templates/applications/)
   to reflect new app URL. For example `http://localhost:8081/v1/api/unsecure/ok`
1. Change all `centralGatewayUrl` to reflect new Application Gateway URL. 
   For example `http://localhost:8080/positive-authorisation/unsecure-always-ok`
1. Deploy all resources (you can omit test Job and Central Gateway, but it's easier to just let them fail)
   on the cluster
1. Build the Mock Application:
   
   <div tabs name="Mock App Build Flavor" group="mock-app-flavor">
   <details open>
   <summary label="dockerized">
   Docker
   </summary>

   ```shell
   export DOCKER_TAG="local"
   export DOCKER_PUSH_REPOSITORY="{Docker username}"
   make image-mock-app
   ```

   </details>
   <details>
   <summary label="local">
   Local
   </summary>

   Change hardcoded application port in [config.go](./tools/external-api-mock-app/config.go), then
   ```shell
   go build ./tools/external-api-mock-app/
   ```
   </details>
   </div>
1. Run the Mock Application:
   
   <div tabs name="Mock App Run Flavor" group="mock-app-flavor">
   <details open>
   <summary label="dockerized">
   Docker
   </summary>

   ```shell
   docker run -p 8180:8080 -p 8190:8090 -v "$PWD/k8s/gateway-test/certs:/etc/secret-volume:ro" "$DOCKER_PUSH_REPOSITORY/mock-app:$DOCKER_TAG"
   ```

   </details>
   <details>
   <summary label="local">
   Local
   </summary>

   ```shell
   ./external-api-mock-app
   ```
   > **CAUTION:** Certificates won't work, unless you copy them from `./k8s/gateway-test/certs` to `/etc/secret-volume`

   </details>
   </div>
1. Run [Central Application Gateway](https://github.com/kyma-project/kyma/tree/main/components/central-application-gateway),
   with `-kubeConfig {path_to_kubeconfig.yaml}` parameter.

You can now send requests to the Application Gateway and debug its behaviour locally.

### Running without cleanup

To run the tests without removing all the resources:

``` shell
make test-gateway-debug
```
