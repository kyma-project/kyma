# Application Connector

## Overview

This is the repository for the Application Connector. It contains the following services:
- Metadata

## Prerequisites

The Application Connector requires Go 1.8 or higher.

## Installation

To install the Application Connector components, follow these steps:

1. `git clone git@github.com/kyma-project/kyma/components/application-connector`
1. `cd application-connector`
1. `make build`

## Usage

This section explains how to use the Application Connector.

### Start the Metadata
To start the Metadata, run this command:

```
./metadata
```

The Metadata has the following parameters:
- **proxyPort** - This port acts as a proxy for the calls from services and lambdas to an external solution. The default port is `8080`.
- **externalAPIPort** - This port exposes the Metadata API to an external solution. The default port is `8081`.
- **eventsTargetURL** - A URL to which you proxy the incoming events. The default URL is http://localhost:9000.
- **appName** - The name of the metadata instance. The default appName is `metadata`.
- **namespace** - Namespace where Metadata is deployed. The default namespace is `kyma-system`.
- **requestTimeout** - A time-out for requests sent through the Metadata. It is provided in seconds. The default time-out is `1`.
- **skipVerify** - A flag for skipping the verification of certificates for the proxy targets. The default value is `false`.
- **requestLogging** - A flag for logging incoming requests. The default value is `false`.

### Sample call

- Creating a new service

```sh
curl -X POST http://localhost:32000/ec-default/v1/metadata/services \
  -d '{"name": "Some EC",
  "provider": "kyma",
  "description": "This is some EC!",
  "api": {
    "targetUrl": "https://ec.com/rest/v2/",
    "credentials": {
      "oauth": {
        "url": "https://ec.com/authorizationserver/oauth/token",
        "clientId": "CLIENT_ID",
        "clientSecret": "CLIENT_SECRET"
      }
    },
    "spec": {
      "apispec": "This is API spec..."
    }
  },
  "events": {
    "spec": {
      "eventsspec": "This is Events Spec..."
    }
  },
  "documentation": {
    "displayName": "Documentation",
    "description": "Description",
    "type": "sometype",
    "tags": ["tag1", "tag2"],
    "docs": [
        {
        "title": "Documentation title...",
        "type": "type",
        "source": "source"
        }
    ]
  }
}'
```

- Fetching all services

```console
curl http://localhost:32000/ec-default/v1/metadata/services
```

## Development

This section explains the development process.

### Rapid development with Telepresence

Application Connector stores its state in the Kubernetes Custom Resource, therefore it's dependent on Kubernetes. You cannot mock the dependency. You cannot develop locally. Manual deployment on every change is a mundane task.

You can, however, leverage [Telepresence](https://www.telepresence.io/). This works by replacing a container in a specified pod, opening up a new local shell or a pre-configured bash, and proxying the network traffic from the local shell through the pod.

Although you are on your local machine, you can make calls such as `curl http://....svc.cluster.local:8081/v1/metadata/services`. When you run a server in this shell, other Kubernetes services can access it.

1. [Install telepresence](https://www.telepresence.io/reference/install).
2. Run Kyma or use the cluster. In the case of Kyma, point your local kubectl to Kyma in Docker.
3. Check the Deployment name to swap and run: `telepresence --namespace <NAMESPACE> --swap-deployment <DEPLOYMENT_NAME>:<CONTAINER_NAME> --run-shell`
```bash
telepresence --namespace kyma-system --swap-deployment metadata:metadata --run-shell
```
4. Every Kubernetes pod has `/var/run/secrets` mounted. The Kubernetes client uses it in the Application Connector services. It is hardcoded. By default, telepresence copies this directory. It stores the directory path in `$TELEPRESENCE_ROOT`, under telepresence shell. It unwinds to `/tmp/tmp...`. You need to move it to `/var/run/secrets`, where the service expects it. Create a symlink:
 ```bash
sudo ln -s $TELEPRESENCE_ROOT/var/run/secrets /var/run/secrets
```
5. Use the `make build` and then the `./metadata` commands, and now all the Kubernetes services that call metadata access this process. The process runs locally on your machine. Use the same command to run different Application Connector services like Proxy or Events.

You can also run another shell to make calls to this service. To run this shell, swap the Remote Environment Broker Deployment, because Istio sidecar is already injected into this Deployment:
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
make test-unit
```

### Generate Kubernetes clients for custom resources

1. Create a directory structure for each client, similar to the one in `pkg/apis`. For example, when generating a client for EgressRule in Istio, the directory structure looks like this: `pkg/apis/istio/v1alpha2`.
2. After creating the directories, define the following files:
    - `doc.go`
    - `register.go`
    - `types.go` - define the custom structs that reflect the fields of the custom resource.

See an example in `pkg/apis/istio/v1alpha2`.

3. Go to the project root directory and run `./hack/update-codegen.sh`. The script generates a new client in `pkg/apis/client/clientset`.


### Contract between the Application Connector and the UI API Facade

The UI API Facade must check the status of the Application Connector services instances that represents the Remote Environment.
In the current solution, the UI API Facade iterates through services to find those which match the criteria, and then uses the health endpoint to determine the status.
The UI API Facade has the following obligatory requirements:
- The Kubernetes service for each Application Connector service uses the `remoteEnvironment` key, with the value as the name of the remote environment.
- The Kubernetes service for each Application Connector service contains one port with the `ext-api-port` name. The system uses this port for the status check.
- Find the Kubernetes Application Connector service service in the `kyma-integration` Namespace. You can change its location in the `ui-api-layer` chart configuration.
- The `/v1/health` endpoint returns a status of `HTTP 200`. Any other status code indicates the service is not healthy.


### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
