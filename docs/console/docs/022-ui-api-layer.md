---
title:  UI API Layer
type: Details
---

UI API Layer is a backend service, which provides API for all views of the Console UI. The service consumes Kubernetes API and exposes simplified GraphQL API to enable frontends to do various K8s resource operations.

## Cache

For read operations on K8s resources, UI API Layer utilizes caching using Informer API. There are separate cache stores per every K8s resource, and they are synchronized during service start. After cache synchronization, a single connection with Kubernetes API server is established, and any event concerning one of observed resources updates the particular cache store. It makes the cache stores always up-to-date without doing multiple requests to Kubernetes API server.

## Modularization

UI API Layer consists of Kubernetes resource logic and cache for different domains, such as Service Catalog,A pplication or Kubeless. The mentioned Kyma components soon will be not required to run Kyma. The UI API Layer introduces modularization changes, which is based on toggling modules while the server is running. Enabled module sychronizes cache for its every resource and enables module logic for all server requests. If a module is disabled, every query, mutation and subscription related to the module returns an error.

Currently, UI API Layer module pluggability is hidden behind a feature toggle. To enable this functionality, run the following command:

```bash
kubectl set env deployment/core-ui-api MODULE_PLUGGABILITY=true -n kyma-system
```

There are following UI API Layer pluggable modules available:
- `apicontroller`
- `authentication`
- `application`
- `content`
- `kubeless`
- `servicecatalog`
- `servicecatalogaddons`

To enable a given module, install proper Helm chart with Kyma component. Inside the chart, there is a `BackendModule` custom resource definition with name equal to the given module.