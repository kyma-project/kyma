---
title: Overview
type: Service Brokers
---

A Service Broker is a server compatible with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification. Each Service Broker registered in Kyma presents the services it offers to the Service Catalog and manages their lifecycle.

The Service Catalog lists all services that the Service Brokers offer. Use the Service Brokers to:
* Provision and deprovision an instance of a service.
* Create and delete a ServiceBinding to link a ServiceInstance to an application.

Each of the Service Brokers available in Kyma performs these operations in a different way. See the documentation of a given Service Broker to learn how it operates.

The Kyma Service Catalog is currently integrated with the following Service Brokers:

* [Application Broker](/components/application-connector#architecture-application-connector-components-application-broker)
* [Helm Broker](/components/helm-broker/#overview-overview)

You can also install these brokers using the Helm Broker's bundles:

* [Azure Service Broker](#service-brokers-azure-service-broker)
* [AWS Service Broker](#service-brokers-aws-broker)

To get the bundles that the Helm Broker provides, go to the [`bundles`](https://github.com/kyma-project/bundles) repository. To build your own Service Broker, follow the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification. For details on how to register a sample Service Broker in the Service Catalog, see [this](#tutorials-register-a-broker-in-the-service-catalog) tutorial.

>**NOTE:** The Service Catalog has the Istio sidecar injected. To enable the communication between the Service Catalog and Service Brokers, either inject Istio sidecar into all brokers or disable mutual TLS authentication.
