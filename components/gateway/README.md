# Gateway

## Overview

This is the repository for the Kyma Gateway.

## Prerequisites

The Gateway requires Go 1.8 or higher.

## Installation

To install the Gateway, follow these steps:

1. `git clone git@github.com:kyma-project/kyma.git`
1. `cd kyma/components/gateway`
1. `CGO_ENABLED=0 go build ./cmd/gateway`

## Usage

This section explains how to use the Gateway.

### Start the Gateway
To start the Gateway, run this command:

```
./gateway
```

The Gateway has the following parameters:
- **proxyPort** - This port acts as a proxy for the calls from services and lambdas to an external solution. The default port is `8080`.
- **externalAPIPort** - This port exposes the Gateway API to an external solution. The default port is `8081`.
- **eventsTargetURL** - A URL to which you proxy the incoming events. The default URL is http://localhost:9000.
- **remoteEnvironment** - Remote Environment name used to write and read information about services. The default remote environment is `default-ec`.
- **appName** - The name of the gateway instance. The default appName is `gateway`.
- **namespace** - Namespace where Gateway is deployed. The default namespace is `kyma-system`.
- **requestTimeout** - A time-out for requests sent through the Gateway. It is provided in seconds. The default time-out is `1`.
- **skipVerify** - A flag for skipping the verification of certificates for the proxy targets. The default value is `false`.
- **requestLogging** - A flag for logging incoming requests. The default value is `false`.
- **proxyTimeout** - A time-out for request send through Gateway proxy in seconds. The default is `10`.
- **proxyCacheTTL** - Time to live of Remote API information stored in proxy cache. The value is provided in seconds and the default is `120`.

The parameters for the Event API correspond to the fields in the [Remote Environment](https://github.com/kyma-project/kyma/tree/master/docs/remote-environment.md):

- **sourceEnvironment** - The name of the Event source environment.
- **sourceType** - The type of the Event source.
- **sourceNamespace** - The organization that publishes the Event.

## Development

This section explains the development process.

### Rapid development with Telepresence

Gateway stores its state in the Kubernetes Custom Resource, therefore Gateway is dependent on Kubernetes. You cannot mock the dependency. You cannot develop locally. Manual deployment on every change is a mundane task.

You can, however, leverage [Telepresence](https://www.telepresence.io/). This works by replacing a container in a specified pod, opening up a new local shell or a pre-configured bash, and proxying the network traffic from the local shell through the pod.

Although you are on your local machine, you can make calls such as `curl http://....svc.cluster.local:8081/v1/health`. When you run a server in this shell, other Kubernetes services can access it.

1. [Install telepresence](https://www.telepresence.io/reference/install).
2. Run Kyma or use the cluster. In the case of Kyma, point your local kubectl to Kyma in Docker.
3. Check the deployment name to swap and run: `telepresence --namespace <NAMESPACE> --swap-deployment <DEPLOYMENT_NAME>:<CONTAINER_NAME> --run-shell`
```bash
telepresence --namespace kyma-system --swap-deployment kyma-core-gateway:gateway --run-shell
```
4. Every Kubernetes pod has `/var/run/secrets` mounted. The Kubernetes client uses it in the Gateway. It is hardcoded. By default, telepresence copies this directory. It stores the directory path in `$TELEPRESENCE_ROOT`, under telepresence shell. It unwinds to `/tmp/tmp...`. You need to move it to `/var/run/secrets`, where the Gateway expects it. Create a symlink:
 ```bash
sudo ln -s $TELEPRESENCE_ROOT/var/run/secrets /var/run/secrets
```
5. Use the `make build` and then the `./gateway` commands, and now all the Kubernetes services that call gateway access this process. The process runs locally on your machine.

You can also run another shell to make calls to this gateway. To run this shell, swap the Remote Environment Broker deployment, because istio sidecar is already injected into this deployment:
```bash
telepresence --namespace kyma-system --swap-deployment kyma-core-remote-environment-broker:reb --run-shell
```

### Generate mocks

To generate a mock, follow these steps:

1. Go to the directory where the interface is located.
2. Run this command:
```sh
mockery -name=Sender
```

### Tests

This section outlines the testing details.

#### Unit tests

To run the unit tests, use the following command:

```
go test `go list ./internal/... ./cmd/...`
```
### Generate Kubernetes clients for custom resources

1. Create a directory structure for each client, similar to the one in `pkg/apis`. For example, when generating a client for EgressRule in Istio, the directory structure looks like this: `pkg/apis/istio/v1alpha2`.
2. After creating the directories, define the following files:
    - `doc.go`
    - `register.go`
    - `types.go` - define the custom structs that reflect the fields of the custom resource.

See an example in `pkg/apis/istio/v1alpha2`.

3. Go to the project root directory and run `./hack/update-codegen.sh`. The script generates a new client in `pkg/apis/client/clientset`.


### Contract between the Gateway and the UI API Facade

The UI API Facade must check the status of the Gateway instance that represents the Remote Environment.
In the current solution, the UI API Facade iterates through services to find those which match the criteria, and then uses the health endpoint to determine the status.
The UI API Facade has the following obligatory requirements:
- The Kubernetes Gateway service uses the `remoteEnvironment` key, with the value as the name of the remote environment.
- The Kubernetes Gateway service contains one port with the `ext-api-port` name. The system uses this port for the status check.
- Find the Kubernetes Gateway service in the `kyma-integration` Namespace. You can change its location in the `ui-api-layer` chart configuration.
- The `/v1/health` endpoint returns a status of `HTTP 200`. Any other status code indicates the service is not healthy.

### Access the Gateway on Minikube

To access the Gateway locally, provide the NodePort of the `core-nginx-ingress-controller` when you send the request.

To get the NodePort, run this command:

```
kubectl -n kyma-system get svc core-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
```

To send the request, run:

```
curl https://gateway.kyma.local:{NODE_PORT}/ec-default/v1/metadata/services --cert ec-default.crt --key ec-default.key -k
```


### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
