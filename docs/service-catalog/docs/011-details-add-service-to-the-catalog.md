---
title: Add a service to the Catalog
type: Details
---

In general, the Service Catalog can expose a service from any Service Broker that is registered in Kyma in accordance with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification.

The Kyma Service Catalog is currently integrated with the following Service Brokers:
* Azure Broker
* Remote Environment Broker
* Helm Broker (experimental)

For details on how to build and register your own Service Broker to expose more services and plans to the cluster users, see the Service Brokers **Overview** document.

>**NOTE:** The Service Catalog has the Istio sidecar injected. To enable the communication between the Service Catalog and Service Brokers, either inject Istio sidecar into all brokers or disable mutual TLS authentication.
