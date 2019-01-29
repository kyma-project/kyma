---
title:  UI API Layer
type: Details
---

The UI API Layer is a backend service which provides an API for all views of the Console UI. This service consumes the Kubernetes API and exposes a simplified GraphQL API to allow frontends to perform Kubernetes resource operations.

## Cache

For GraphQL queries, the UI API Layer uses caching which is based on Informers from the Kubernetes Go client. There are separate cache stores for every Kubernetes resource. The stores are synchronized when the service starts. After cache synchronization, a single connection with the Kubernetes API server is established and any event related to one of the observed resources updates the given cache store. This logic ensures that cache stores are always up-to-date without sending multiple requests to the Kubernetes API server.

## Modularization

The UI API Layer consists of the Kubernetes resource logic and cache for different domains, such as the Service Catalog, Application, or Kubeless. The UI API Layer introduces modularization changes which are based on toggling modules while the server is running. The enabled module sychronizes cache for its resource and enables the module's logic for all server requests. If you disable a given module, every GraphQL query, mutation, and subscription related to this module returns an error.

The UI API Layer module pluggability is hidden behind a feature toggle. It is not enabled by default because the Console UI still requires resiliency improvements to ensure no errors occur when a certain Kyma component is not installed.

To enable this functionality, run the following command:

```bash
kubectl set env deployment/core-ui-api MODULE_PLUGGABILITY=true -n kyma-system
```

These are the available UI API Layer pluggable modules which contain the GraphQL resolver logic, where:
- `apicontroller` relates to the API Controller.
- `authentication` relates to IDP Presets.
- `application` relates to the Application Connector.
- `content` relates to the documentation.
- `kubeless` relates to serverless.
- `servicecatalog` relates to the Service Catalog, including Service Classes, Service Instances, and Service Bindings.
- `servicecatalogaddons` relates to the Service Catalog add-ons, such as ServiceBindingUsage, and UsageKinds.

To enable a given module, install a proper Kyma component. It includes the BackendModule custom resource definition with the same name as the name of a given module.