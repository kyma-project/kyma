---
title: Service Brokers
type: Overview
---

A Service Broker is a server compatible with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification. A Service Broker manages the lifecycle of one or more services. It advertises a catalog of service offerings and service plans to a platform.

The Service Catalog lists all services that the Service Brokers offer. Use the Service Brokers to:
* Provision and de-provision an instance of a service
* Create and delete a service binding

Create a service binding to link a service instance to an application. During this process, credentials are delivered in Secrets to provide you with the information necessary to connect to the service instance. The process of deleting a service binding is known as unbinding.

Each of the Service Brokers available in Kyma performs these operations in a different way. See the documentation for a given Service Broker to learn how it operates.

Kyma provides these Service Brokers to use with the Service Catalog:

* Azure Service Broker
* Google Cloud Platform Service Broker (Experimental)
* Application Broker
* Helm Broker

Follow the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification to build your own Service Broker.
Register every new Service Broker in the Service Catalog to make the services and plans available to the users. For more information on registering Service Brokers in the Service Catalog, see the [Service Catalog Demonstration Walkthrough](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/walkthrough.md).
