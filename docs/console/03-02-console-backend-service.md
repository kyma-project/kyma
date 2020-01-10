---
title: Console Backend Service
type: Details
---

The Console Backend Service is a backend service which provides an API for all views of the Console UI. This service consumes the Kubernetes API and exposes a simplified GraphQL API to allow frontends to perform Kubernetes resource operations.

> **NOTE:** Read [this](/components/security#details-graph-ql) security document for more information about the Kyma GraphQL implementation.

## Cache

For GraphQL queries, the Console Backend Service uses caching which is based on Informers from the Kubernetes Go client. There are separate cache stores for every Kubernetes resource. The stores are synchronized when the service starts. After cache synchronization, a single connection with the Kubernetes API server is established and any event related to one of the observed resources updates the corresponding cache store. This logic ensures that cache stores are always up-to-date without sending multiple requests to the Kubernetes API server.

## Modularization

The Console Backend Service consists of the Kubernetes resource logic and cache for different domains, such as the Service Catalog, Application, or Kubeless. The Console Backend Service introduces modularization changes which are based on toggling modules while the server is running. The enabled module synchronizes cache for its resource and enables the module's logic for all server requests. If you disable a given module, every GraphQL query, mutation, and subscription related to this module returns an error.

These are the available Console Backend Service pluggable modules which contain the GraphQL resolver logic, where:
- `apicontroller` relates to the API Controller.
- `apigateway` relates to the API Gateway.
- `authentication` relates to IDP Presets.
- `application` relates to the Application Connector.
- `kubeless` relates to Serverless.
- `rafter` relates to Rafter.
- `servicecatalog` relates to the Service Catalog, including Service Classes, Service Instances, and Service Bindings.
- `servicecatalogaddons` relates to the Service Catalog add-ons, such as ServiceBindingUsage, and UsageKinds.
- `grafana` relates to Grafana.
- `loki` relates to Loki.

To enable a given module, install the corresponding Kyma component. It includes the BackendModule custom resource with the same name as the name of a given module.
