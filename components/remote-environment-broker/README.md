# Remote Environment Broker

## Overview

The Remote Environment Broker (REB) provides remote environments in the [Service Catalog](../../docs/service-catalog/docs/001-overview-service-catalog.md).
A remote environment represents the environment connected to the Kyma instance.
The REB implements the [Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).

The REB fetches all the remote environments' custom resources and exposes their APIs and Events as service classes to the Service Catalog.
When the remote environments list is available in the Service Catalog, you can provision those service classes and enable Kyma services to use them.

For more details about provisioning, deprovisioning, binding, and unbinding, see the [Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) documentation.

## Prerequisites

You need the following tools to set up the project:
* The 1.9 or higher version of [Go](https://golang.org/dl/)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Dep](https://github.com/golang/dep)

## Development

Before each commit, use the `before-commit.sh` script, which tests your changes.

## Code generation

Structs related to Custom Resource Definitions are defined in `pkg/apis/remoteenvironment/v1alpha1/types.go` and registered in `pkg/apis/remoteenvironment/v1alpha1/`. After making any changes there, please run:
```bash
./hack/update-codegen.sh
```
