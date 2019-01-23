# Application Broker

## Overview

The Application Broker (AB) provides applications in the [Service Catalog](../../docs/service-catalog/docs/01-01-service-catalog.md).
An Application represents a remote application connected to the Kyma instance.
The AB implements the [Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).

The AB fetches all the applications' custom resources and exposes their APIs and Events as service classes to the Service Catalog.
When the applications list is available in the Service Catalog, you can provision those service classes and enable Kyma services to use them.

The AB works as a Namespace-scoped broker which is registered in the specific Namespace when the ApplicationMapping is created in this Namespace.

For more details about provisioning, deprovisioning, binding, and unbinding, see the [Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) documentation.

## Prerequisites

You need the following tools to set up the project:
* The 1.9 or higher version of [Go](https://golang.org/dl/)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Dep](https://github.com/golang/dep)

## Development

Before each commit, use the `before-commit.sh` script or the `make build` command to test your changes.

### Use environment variables

| Name | Required | Default | Description |
|-----|---------|--------|------------|
|**APP_PORT** | NO | `8080` | The port on which the HTTP server listens |
|**APP_BROKER_RELIST_DURATION_WINDOW** | YES | - | Time period after which the AB synchronizes with the Service Catalog if a new Application is added. In case more than one Application is added, synchronization is performed only once. |
| **APP_UNIQUE_SELECTOR_LABEL_KEY** | YES | - | Defined label key selector which allows uniquely identify AB pod's |
| **APP_UNIQUE_SELECTOR_LABEL_VALUE** | YES | - | Defined label value selector which allows uniquely identify AB pod's |
| **NAMESPACE** | YES | - | AB working Namespace |


## Code generation

Structs related to CustomResourceDefinitions are defined in `pkg/apis/application/v1alpha1/types.go` and registered in `pkg/apis/application/v1alpha1/`. After making any changes there, please run:
```bash
./hack/update-codegen.sh
```
