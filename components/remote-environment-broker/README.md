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

## Details

The Remote Environment Broker converts the RemoteEnvironment Custom Resource to the Open Service Broker API [Service Object](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#service-objects).
All fields are mapped one-to-one except from the **name** and **metadata.providerDisplayName** properties.  

The pattern for the **metadata.providerDisplayName** property is as follows:
```
{RemoteEnvironment.Service.ProviderDisplayName} - {RemoteEnvironment.Name}
```
It enables you to easily distinguish different remote environments from the same provider. 

The pattern for the **name** property is as follows:
```
{NORMALIZED_DISPLAY_NAME}-{FIST_FIVE_CHARARACTERS_OF_SHA_FROM_SERVICE_ID}
```
After normalization, the **displayName** contains only lowercase characters, numbers and hyphens.

## Development

Before each commit, use the `before-commit.sh` script, which tests your changes.

### Use environment variables
Remote Environment Broker can work in 2 modes: as a cluster-scoped or namespace-scoped broker (see **APP_CLUSTER_SCOPED_BROKER_ENABLED**).
In the first case, on creation of a Remote Environment, proper Cluster Service Classes will be created automatically and will be available for every namespace.
In the second case, on creation of a Environment Mapping, namespaced-scoped broker is registered in the specific namespace and  services from this Remote Environment are visible in the namespace.

| Name | Required | Default | Description |
|-----|---------|--------|------------|
|**APP_PORT** | false | 8080 | The port on which the HTTP server listen. | 
|**APP_BROKER_RELIST_DURATION_WINDOW** | true | - | Synchronize REB in Service Catalog (if needed) at most once per this period |
| **APP_CLUSTER_SCOPED_BROKER_NAME**| true | - | Name of the ClusterServiceBroker (if registered as a cluster-scoped broker) |
| **APP_CLUSTER_SCOPED_BROKER_ENABLED** | true | - | Flag defines if REB is working as a ClusterServiceBroker or a ServiceBroker | 
| **APP_UNIQUE_SELECTOR_LABEL_KEY** | true | - |  APP_UNIQUE_SELECTOR_LABEL_KEY and APP_UNIQUE_SELECTOR_LABEL_VALUE define label selector which uniquely identify REB pod's |
| **APP_UNIQUE_SELECTOR_LABEL_VALUE** | true | - | see above |
| **NAMESPACE** | true | - | REB working namespace |
  
 
## Code generation

Structs related to Custom Resource Definitions are defined in `pkg/apis/remoteenvironment/v1alpha1/types.go` and registered in `pkg/apis/remoteenvironment/v1alpha1/`. After making any changes there, please run:
```bash
./hack/update-codegen.sh
```
