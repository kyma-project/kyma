# Application Broker

## Overview

The Application Broker (AB) provides applications in the [Service Catalog](../../docs/service-catalog/01-01-service-catalog.md).
An Application represents a remote application connected to the Kyma instance.
The AB implements the [Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).

The AB fetches all the applications' custom resources and exposes their APIs and Events as service classes to the Service Catalog.
When the applications list is available in the Service Catalog, you can provision those service classes and enable Kyma services to use them.

The AB works as a Namespace-scoped broker which is registered in the specific Namespace when the ApplicationMapping is created in this Namespace.

For more details about provisioning, deprovisioning, binding, and unbinding, see the [Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) documentation.

## Prerequisites

To set up the project, download these tools:

* [Go](https://golang.org/dl/) 1.11.4
* [Dep](https://github.com/golang/dep) v0.5.0
* [Docker](https://www.docker.com/)

These Go and Dep versions are compliant with the `buildpack` used by Prow. For more details read [this](https://github.com/kyma-project/test-infra/blob/master/prow/images/buildpack-golang/README.md) document.

## Development

Before each commit, use the `make verify` command to test your changes. To build an image, use the `make build-image` command with **DOCKER_PUSH_REPOSITORY** and **DOCKER_PUSH_DIRECTORY** variables, for example:
```
DOCKER_PUSH_REPOSITORY=eu.gcr.io DOCKER_PUSH_DIRECTORY=/kyma-project/develop make build-image
```

### Use environment variables

| Name | Required | Default | Description |
|-----|---------|--------|------------|
|**APP_PORT** | NO | `8080` | The port on which the HTTP server listens |
|**APP_BROKER_RELIST_DURATION_WINDOW** | YES | None | Time period after which the AB synchronizes with the Service Catalog if a new Application is added. In case more than one Application is added, synchronization is performed only once. |
| **APP_SERVICE_NAME** | YES | None | The name of the Kubernetes service which exposes the Service Brokers API |
| **APP_UNIQUE_SELECTOR_LABEL_KEY** | YES | None | Defined label key selector which allows uniquely identify AB pod's |
| **APP_UNIQUE_SELECTOR_LABEL_VALUE** | YES | None | Defined label value selector which allows uniquely identify AB pod's |
| **NAMESPACE** | YES | None | AB working Namespace |


## Code generation

Structs related to CustomResourceDefinitions are defined in `pkg/apis/application/v1alpha1/types.go` and registered in `pkg/apis/application/v1alpha1/`. After making any changes there, please run:

```bash
./hack/update-codegen.sh
```
