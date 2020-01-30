# Application Registry

This is the repository for the Application Registry, which is a part of the Application Connector. See [this](../../docs/application-connector/) directory for more documentation for this component.

## Prerequisites

The Application Registry requires Go 1.8 or higher.

## Installation

To install the Application Registry components, follow these steps:

1. `git clone git@github.com/kyma-project/kyma.git`
2. `cd components/application-registry`
3. `CGO_ENABLED=0 go build ./cmd/applicationregistry`

## Usage

This section explains how to use the Application Registry.

### Start the Application Registry

To start the Application Registry, run this command:

```
./applicationregistry
```

The Application Registry has the following parameters:
- **externalAPIPort** is the port that exposes the Metadata API to an external solution. The default port is `8081`.
- **proxyPort** is the port that acts as a proxy for the calls from services and lambdas to an external solution. The default port is `8080`.
- **namespace** is the Namespace where Application Registry is deployed. The default Namespace is `kyma-system`.
- **requestTimeout** is the time-out for requests sent through the Application Registry. It is provided in seconds. The default time-out is `1`.
- **requestLogging** is the flag for logging incoming requests. The default value is `false`.
- **specRequestTimeout** is the time-out for requests fetching specifications provided by the user. It is provided in seconds. The default time-out is `20`.
- **rafterRequestTimeout** is the time-out for requests fetching specifications from Rafter. It is provided in seconds. The default time-out is `20`.
- **detailedErrorResponse** is the flag for showing detailed internal error messages in response bodies. The default value is `false` and all internal server error messages are shortened to `Internal error`, while all other error messages are shown as usual.
- **uploadServiceURL** is the URL of the Upload Service.
- **insecureAssetDownload** is the flag for skipping certificate verification for asset download. The default value is `false`.
- **insecureSpecDownload** is the flag for skipping certificate verification for API specification download. The default value is `false`.

### Sample calls

- Create a new service

```sh
curl -X POST https://gateway.kyma.local/{APPLICATION_NAME}/v1/metadata/services --cert {CER_NAME}.crt --key {CERT_KEY}.key -k \
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

- Fetch all services

```
curl https://gateway.kyma.local/{APPLICATION_NAME}/v1/metadata/services --cert {CERT_NAME}.crt --key {KEY_NAME}.key -k
```

## Development

This section explains the development process.

### Rapid development with Telepresence

The Application Registry stores its state in the Kubernetes custom resource, therefore depends on Kubernetes. You cannot mock the dependency. You cannot develop locally. Manual deployment on every change is a mundane task.

You can, however, leverage [Telepresence](https://www.telepresence.io/). This works by replacing a container in a specified Pod, opening up a new local shell or a pre-configured bash, and proxying the network traffic from the local shell through the Pod.

Although you are on your local machine, you can make calls such as `curl http://....svc.cluster.local:8081/v1/metadata/services`. When you run a server in this shell, other Kubernetes services can access it.

1. [Install telepresence](https://www.telepresence.io/reference/install).
2. Run Kyma or use the cluster. In the case of Kyma, point your local kubectl to Kyma in Docker.
3. Check the Deployment name to swap and run: `telepresence --namespace <NAMESPACE> --swap-deployment <DEPLOYMENT_NAME>:<CONTAINER_NAME> --run-shell`
```bash
telepresence --namespace kyma-system --swap-deployment metadata:metadata --run-shell
```
4. Every Kubernetes Pod has `/var/run/secrets` mounted. The Kubernetes client uses it in the Application Connector services. It is hardcoded. By default, telepresence copies this directory. It stores the directory path in `$TELEPRESENCE_ROOT`, under telepresence shell. It unwinds to `/tmp/tmp...`. You need to move it to `/var/run/secrets`, where the service expects it. Create a symlink:
 ```bash
sudo ln -s $TELEPRESENCE_ROOT/var/run/secrets /var/run/secrets
```
5. Run `CGO_ENABLED=0 go build ./cmd/applicationregistry` to build the Application Registry and give all Kubernetes services that call the Application Registry access to this process. The process runs locally on your machine. Use the same command to run different Application Connector services like Proxy or Events.

You can also run another shell to make calls to this service. To run this shell, swap the Application Broker Deployment because Istio sidecar is already injected into this Deployment:
```bash
telepresence --namespace kyma-system --swap-deployment kyma-core-application-broker:reb --run-shell
```

### Generate mocks

To generate a mock, follow these steps:

1. Go to the directory where the interface is located.
2. Run this command:
```sh
mockery -name=Sender
```

## Tests

This section outlines the testing details.

### Run tests locally

When you develop the Application Connector components, you can test the changes you introduced on a local Kyma deployment before you push them to a production cluster.
To test the component you modified, run the `run-with-local-tests.sh` script located in the `scripts` directory.

Running the script builds the Docker image of the component, pushes it to the Minikube registry, and updates the component deployment in the Minikube cluster. It then triggers the `run-local-tests.sh` script, which builds the image of the acceptance tests to the Minikube registry, creates a Pod with the tests, and fetches the logs from that Pod.

Alternatively, you can run only the `run-local-tests.sh` script for the given component to build the image of the component's acceptance tests to the Minikube registry, create a Pod with the tests, and fetch the logs from that Pod.

### Unit tests

To run the unit tests, use the following command:

```
go test ./...
```

## Generate Kubernetes clients for custom resources

1. Create a directory structure for each client, similar to the one in `pkg/apis`. For example, when generating a client for EgressRule in Istio, the directory structure looks like this: `pkg/apis/istio/v1alpha2`.
2. After creating the directories, define the following files:
    - `doc.go`
    - `register.go`
    - `types.go` - define the custom structs that reflect the fields of the custom resource.

See an example in `pkg/apis/istio/v1alpha2`.

3. Go to the project root directory and run `./hack/update-codegen.sh`. The script generates a new client in `pkg/apis/client/clientset`.

## Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
