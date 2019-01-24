---
title: Add a service to the Catalog
type: Details
---

In general, the Service Catalog can expose a service from any Service Broker that is registered in Kyma in accordance with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification.

The Kyma Service Catalog is currently integrated with the following Service Brokers:
* Azure Broker
* Application Broker
* Helm Broker (experimental)
* GCP Broker

For details on how to register a sample Service Broker, see the Service Brokers [tutorial](#tutorials-register-a-broker-in-the-service-catalog).

>**NOTE:** The Service Catalog has the Istio sidecar injected. To enable the communication between the Service Catalog and Service Brokers, either inject Istio sidecar into all brokers or disable mutual TLS authentication.
