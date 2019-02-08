---
title: Service Brokers
type: Service Brokers
---

A Service Broker is a server compatible with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification. Each Service Broker registered in Kyma presents the services it offers to the Service Catalog and manages their lifecycle.

The Service Catalog lists all services that the Service Brokers offer. Use the Service Brokers to:
* Provision and de-provision an instance of a service
* Create and delete a service binding

Create a service binding to link a service instance to an application. During this process, credentials are delivered in Secrets to provide you with the information necessary to connect to the instance of a service. The process of deleting a service binding is known as unbinding. Each of the Service Brokers available in Kyma performs these operations in a different way. See the documentation of a given Service Broker to learn how it operates.

The Kyma Service Catalog is currently integrated with the following Service Brokers:

* [Application Broker](/components/application-connector#architecture-application-connector-components-application-broker)
* [Helm Broker](/components/helm-broker#overview-helm-broker)

Moreover, you can install these brokers using the Helm Broker's bundles:

* [Google Cloud Platform (GCP) Broker](#service-brokers-gcp-broker)
* [Azure Service Broker](#service-brokers-azure-service-broker)

To get the bundles that the Helm Broker provides, go to the [`bundles`](https://github.com/kyma-project/bundles) repository. To build your own Service Broker, follow the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification. For details on how to register a sample Service Broker in the Service Catalog, see [this](#tutorials-register-a-broker-in-the-service-catalog) tutorial.

>**NOTE:** The Service Catalog has the Istio sidecar injected. To enable the communication between the Service Catalog and Service Brokers, either inject Istio sidecar into all brokers or disable mutual TLS authentication.
