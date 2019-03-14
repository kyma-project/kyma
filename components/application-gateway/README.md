# Application Gateway

## Overview

This is the repository for the Kyma Application Gateway.

## Prerequisites

The Application Gateway requires Go 1.8 or higher.

## Installation

To install the Application Gateway, follow these steps:

1. `git clone git@github.com:kyma-project/kyma.git`
1. `cd kyma/components/application-gateway`
1. `CGO_ENABLED=0 go build ./cmd/applicationgateway`

## Usage

This section explains how to use the Application Gateway.

### Start the Application Gateway
To start the Application Gateway, run this command:

```
./applicationgateway
```

The Application Gateway has the following parameters:
- **proxyPort** - This port acts as a proxy for the calls from services and lambdas to an external solution. The default port is `8080`.
- **externalAPIPort** - This port exposes the API allowing to check component status. The default port is `8081`.
- **remoteEnvironment** - Remote Environment name used to write and read information about services. The default remote environment is `default-ec`.
- **namespace** - Namespace where Application Gateway is deployed. The default namespace is `kyma-system`.
- **requestTimeout** - A timeout for requests sent through the Application Gateway. It is provided in seconds. The default time-out is `1`.
- **skipVerify** - A flag for skipping the verification of certificates for the proxy targets. The default value is `false`.
- **requestLogging** - A flag for logging incoming requests. The default value is `false`.
- **proxyTimeout** - A timeout for request send through proxy in seconds. The default is `10`.
- **proxyCacheTTL** - Time to live of Remote API information stored in proxy cache. The value is provided in seconds and the default is `120`.

## Development

This section explains the development process.

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


### Contract between the Application Gateway and the Console Backend Service

The Console Backend Service must check the status of the Application Gateway instance that represents the Remote Environment.
In the current solution, the Console Backend Service iterates through services to find those which match the criteria, and then uses the health endpoint to determine the status.
The Console Backend Service has the following obligatory requirements:
- The Kubernetes service uses the `application` key, with the value as the name of the remote environment.
- The Kubernetes service contains one port with the `http-api-port` name. The system uses this port for the status check.
- Find the Kubernetes service in the `kyma-integration` Namespace. You can change its location in the `console-backend-service` chart configuration.
- The `/v1/health` endpoint returns a status of `HTTP 200`. Any other status code indicates the service is not healthy.

### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
