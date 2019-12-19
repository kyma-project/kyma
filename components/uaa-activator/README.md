# UAA Activator

## Overview

The UAA Activator activates the UAA that manages application authorizations and trust to identity providers. It configures the whole Kyma environment to support the UAA authorization when logging to the Console UI.
The UAA Activator application is also desired to be executed multiple times as it ensures that clusters are in a desired state.

## Prerequisites

To set up the project, download these tools:

* [Go](https://golang.org/dl/) v1.13 or higher
* [Dep](https://github.com/golang/dep) v0.5.0
* [Docker](https://www.docker.com/) in the newest version

## Usage

This section explains how to use the UAA Activator.

### Run a local version

To run the application without building the binary file, run this command:

```bash
export UAA_SERVICE_INSTANCE_PARAMS_SECRET_NAME="uaa-activator"
export UAA_SERVICE_INSTANCE_PARAMS_SECRET_KEY="security.json"
export UAA_SERVICE_INSTANCE_NAMESPACE="kyma-system"
export UAA_SERVICE_INSTANCE_NAME="uaa-issuer"
export UAA_SERVICEBINDING_NAMESPACE="kyma-system"
export UAA_SERVICEBINDING_NAME="uaa-issuer-secret"

export CLUSTER_DOMAIN_NAME={your_cluster_domain_name}

export DEX_CONFIG_MAP_NAME="dex-config"
export DEX_CONFIG_MAP_NAMESPACE="kyma-system"
export DEX_DEPLOYMENT_NAME="dex"
export DEX_DEPLOYMENT_NAMESPACE="kyma-system"

export GLOBAL_REPEAT_INTERVAL="1s"
export GLOBAL_REPEAT_TIMEOUT="5m"

export DEVELOPMENT_LOGGER="true"
go run main.go
```

This is the example execution output:
```
{"level":"info","ts":1576762653.0310497,"caller":"scheduler/scheduler.go:48","msg":"Started execution of 6 steps...","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8"}
{"level":"info","ts":1576762653.0311215,"caller":"scheduler/scheduler.go:50","msg":"Waiting until the UAA class and plan definition are available","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"start"}
{"level":"info","ts":1576762653.03841,"caller":"scheduler/scheduler.go:57","msg":"Waiting until the UAA class and plan definition are available","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"done"}
{"level":"info","ts":1576762653.0384443,"caller":"scheduler/scheduler.go:50","msg":"Provisioning and waiting for ready UAA instance","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"start"}
{"level":"info","ts":1576762655.0570104,"caller":"scheduler/scheduler.go:57","msg":"Provisioning and waiting for ready UAA instance","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"done"}
{"level":"info","ts":1576762655.0570755,"caller":"scheduler/scheduler.go:50","msg":"Creating and waiting for ready binding for the UAA instance","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"start"}
{"level":"info","ts":1576762656.1846604,"caller":"scheduler/scheduler.go:57","msg":"Creating and waiting for ready binding for the UAA instance","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"done"}
{"level":"info","ts":1576762656.1847034,"caller":"scheduler/scheduler.go:50","msg":"Creating Dex override with the UAA connector (used later for Kyma upgrade)","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"start"}
{"level":"info","ts":1576762656.1938248,"caller":"scheduler/scheduler.go:57","msg":"Creating Dex override with the UAA connector (used later for Kyma upgrade)","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"done"}
{"level":"info","ts":1576762656.193866,"caller":"scheduler/scheduler.go:50","msg":"Updating current Dex ConfigMap with UAA connector entry","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"start"}
{"level":"info","ts":1576762656.2050023,"caller":"scheduler/scheduler.go:57","msg":"Updating current Dex ConfigMap with UAA connector entry","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"done"}
{"level":"info","ts":1576762656.2050426,"caller":"scheduler/scheduler.go:50","msg":"Updating current Dex Deployment to use UAA connector and waiting for ready state","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"start"}
{"level":"info","ts":1576762676.2660787,"caller":"scheduler/scheduler.go:57","msg":"Updating current Dex Deployment to use UAA connector and waiting for ready state","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","action":"done"}
{"level":"info","ts":1576762676.266128,"caller":"scheduler/scheduler.go:59","msg":"All steps executed without errors.","execution_id":"c6ef3ed0-e360-4c15-b4ca-131c296c6cb8","total_execution_time":"23.235080807s","step_executed":6}
```

>**NOTE:** The **execution_id** is generated for every new execution. You can use that for debigging to get logs from the given execution. 


### Use environment variables

Use the following environment variables to configure the application:
	
| Name                                        | Required | Default                                                           | Description                                                                                                                                                             |
|---------------------------------------------|:--------:|-------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **UAA_SERVICE_INSTANCE_PARAMS_SECRET_NAME** | Yes      | None                                                              | Specifies the Secret name that contains parameters that are passed during the provisioning of the UAA ServiceInstance.                                                      |
| **UAA_SERVICE_INSTANCE_PARAMS_SECRET_KEY**  | Yes      | None                                                              | Specifies the Secret key name under which the parameters are stored. Each parameter must be a JSON object.                                                              |
| **UAA_SERVICE_INSTANCE_NAMESPACE**          | Yes      | None                                                              | Specifies the Namespace where the ServiceInstance is created by the UAA Activator.                                                                             |
| **UAA_SERVICE_INSTANCE_NAME**               | Yes      | None                                                              | Specifies the name of the ServiceInstance that is created by the UAA Activator.                                                                                |
| **UAA_SERVICEBINDING_NAMESPACE**            | Yes      | None                                                              | Specifies the Namespace where the ServiceBinding is created by the UAA Activator.                                                                              |
| **UAA_SERVICEBINDING_NAME**                 | Yes      | None                                                              | Specifies the name of the ServiceBinding that is created by the UAA Activator.                                                                                 |
| **UAA_CLUSTER_SERVICE_CLASS_NAME**          | No       | `xsuaa`                                                           | Specifies the name of the UAA ClusterServiceClass.                                                                                                                      |
| **UAA_CLUSTER_SERVICE_PLAN_NAME**           | No       | `z54zhz47zdx5loz51z6z58zhvcdz59-b207b177b40ffd4b314b30635590e0ad` | Specifies the name of the UAA ClusterServicePlan.                                                                                                                       |
| **CLUSTER_DOMAIN_NAME**                     | Yes      | None                                                              | Specifies the domain name of the cluster where the UAA Activator is deployed.                                                                                           |
| **DEX_CONFIG_MAP_NAME**                     | Yes      | None                                                              | Specifies the name of the ConfigMap that holds the Dex configuration.                                                                                                   |
| **DEX_CONFIG_MAP_NAMESPACE**                | Yes      | None                                                              | Specifies the Namespace of the ConfigMap that holds the Dex configuration.                                                                                              |
| **DEX_DEPLOYMENT_NAME**                     | Yes      | None                                                              | Specifies the name of the Dex Deployment.                                                                                                                               |
| **DEX_DEPLOYMENT_NAMESPACE**                | Yes      | None                                                              | Specifies the Namespace where the Dex Deployment is available.                                                                                                          |
| **GLOBAL_REPEAT_INTERVAL**                  | No       | `1s`                                                              | Specifies the time interval after which the failed operation is repeated. The valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.                               |
| **GLOBAL_REPEAT_TIMEOUT**                   | No       | `5m`                                                              | Specifies the maximum time during which the failed operation is being repeated. The valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.                         |
| **DEVELOPMENT_LOGGER**                      | No       | `false`                                                           | Specifies whether to use the development logger that writes logs for the `debug` level and the levels above to standard the error output stream in a human-friendly format. |

## Development

Use the `make verify` command to test your changes before each commit. To build an image, use the `make build-image` command with the **DOCKER_PUSH_REPOSITORY** and **DOCKER_PUSH_DIRECTORY** variables, for example:
```bash
DOCKER_PUSH_REPOSITORY=eu.gcr.io DOCKER_PUSH_DIRECTORY=/kyma-project/ make build-image
```

## Known issues

Currently, the UAA ServiceInstance is not updated when the content from the **UAA_SERVICE_INSTANCE_PARAMS_SECRET_NAME** Secret is changed while the Secret name remains the same.
