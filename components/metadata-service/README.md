# Metadata Service

This is the repository for the Metadata Service, which is a part of the Application Connector. See [this](../../docs/application-connector/docs/) directory for more documentation for this component.

## Prerequisites

The Metadata Service requires Go 1.8 or higher.

## Installation

To install the Metadata Service components, follow these steps:

1. `git clone git@github.com/kyma-project/kyma/components/metadata-service`
2. `cd components/metadata-service`
3. `CGO_ENABLED=0 go build ./cmd/metadataservice`

## Usage

This section explains how to use the Metadata Service.

### Start the Metadata service

To start the Metadata Service, run this command:

```
./metadataservice
```

The Metadata Service has the following parameters:
- **proxyPort** - This port acts as a proxy for the calls from services and lambdas to an external solution. The default port is `8080`.
- **externalAPIPort** - This port exposes the Metadata API to an external solution. The default port is `8081`.
- **minioURL** - The URL of a Minio service which stores specifications and documentation.
- **namespace** - Namespace where Metadata Service is deployed. The default Namespace is `kyma-system`.
- **requestTimeout** - A time-out for requests sent through the Metadata Service. It is provided in seconds. The default time-out is `1`.
- **requestLogging** - A flag for logging incoming requests. The default value is `false`.
- **detailedErrorResponse** - A flag for showing detailed internal error messages in response bodies. The default value is `false` and all internal server error messages are shortened to `Internal error`, while all other error messages are shown as usual.

### Sample calls

To access the Metadata Service on Minikube you need the NodePort of `application-connector-nginx-ingress-controller`.
To get the NodePort, run this command:

```
kubectl -n kyma-system get svc application-connector-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
```

- Create a new service 

```sh
curl -X POST https://gateway.kyma.local:{NODE_PORT}/ec-default/v1/metadata/services --cert ec-default.crt --key ec-default.key -k \
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
curl https://gateway.kyma.local:{NODE_PORT}/ec-default/v1/metadata/services --cert ec-default.crt --key ec-default.key -k
```

## Development

This section explains the development process.

### Rapid development with Telepresence

The Metadata Service stores its state in the Kubernetes Custom Resource, therefore it's dependent on Kubernetes. You cannot mock the dependency. You cannot develop locally. Manual deployment on every change is a mundane task.

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
5. Run `CGO_ENABLED=0 go build ./cmd/metadata` to build the Metadata Service and give all  Kubernetes services that call the Metadata Service access to this process. The process runs locally on your machine. Use the same command to run different Application Connector services like Proxy or Events.

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
go test ./...
```

### Generate Kubernetes clients for custom resources

1. Create a directory structure for each client, similar to the one in `pkg/apis`. For example, when generating a client for EgressRule in Istio, the directory structure looks like this: `pkg/apis/istio/v1alpha2`.
2. After creating the directories, define the following files:
    - `doc.go`
    - `register.go`
    - `types.go` - define the custom structs that reflect the fields of the custom resource.

See an example in `pkg/apis/istio/v1alpha2`.

3. Go to the project root directory and run `./hack/update-codegen.sh`. The script generates a new client in `pkg/apis/client/clientset`.

### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
