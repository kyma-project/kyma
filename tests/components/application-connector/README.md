# Component tests of Application Connector

## Tests structure 

Test are structured as a monorepo with tests for modules:
- Application Gateway

## Application Gateway tests

Tests are executed on a Kyma cluster where tested Application Gateway is installed. 

The environment consists of Kubernetes pod running tests and Mock-Application simulating remote endpoints for tested Application Gateway.

<!---TODO: Draw a simple diagram here--->

Test cases are defined as services in Application CRD.

The whole test setup is deployed into Kyma cluster with helm template executed in the Makefile by the command `test-gateway`.  
 
Image versions, and external service name used during testing can be set up in Helm chart values file `k8s/gateway-test/values.yaml` 

### Local Build of test images

1. Setup DOCKER_TAG and DOCKER_PUSH_REPOSITORY in `local_build.sh` for your target Docker registry settings
2. Run `local_build.sh`

### Local execution

1. Provision Kubernetes local cluster with k3d

```shell
kyma provision k3d
```
2. Install minimal set of components required to run Application Gateway with specific variant

For Application Gateway deployed as part of KymaOS cluster run:

```shell
kyma deploy --components-file mini-kyma-os.yaml 
```

For Application Gateway deployed as part of Managed SAP Kyma Runtime run:

```shell
kyma deploy --components-file mini-kyma-skr.yaml 
```

3. Run command `make test-gateway`






