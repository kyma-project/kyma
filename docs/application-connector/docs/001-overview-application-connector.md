---
title: Overview
---

The Application Connector is a part of the Kyma project which simplifies integration of various systems with Kyma.

No matter if you want to integrate an on-premise or a cloud system, the Application Connector ensures that integration works in the same way and developers are not distracted by the configuration or networking issues.

The Application Connector ensures that a connected system communicates with Kyma securely using the client certificate which can be acquired from the Connector service. A client certificate ensures that each connected system is separated.

The Application Connector provides a possibility to communicate with services and lambdas deployed in Kyma in an asynchronous matter using events. A system can send an event which triggers a subscribed service or lambda. Developers can benefit from the built-in monitoring and tracing which allow troubleshooting of the event delivery.

The Application Connector integrates a connected system with a Service Catalog. All APIs and all Events which are available in the system can be registered using the Metadata service. The registration process integrates all components into the Service Catalog. Developers can browser documentation of the registered APIs and Event Catalog and can control the access to it. [Better description required]

The Application Connector Proxy service is tunneling requests to the system's API. Developers don't need to take care of the endpoint configuration. The Proxy service is also able to automate the security token handling. The API can be registered together with client credentials, and OAuth token will be acquired and cached automatically.

All components of the Application Connector can be scaled independently to adjust to the need of the solution which is being built on the top of the Kyma.


## Functionality

The Application Connector brings the following functionality to Kyma:

- Manage the lifecycle of the Remote Environment which is representing the connected system.
- Establishing secured connection and generation of the client certificate used by the system which is connecting to Kyma.
- Register of the external system APIs and Event catalogs.
- Deliver of events from the external system to Kyma Eventbus.
- Proxing calls from the Kyma to external APIs registered by connected systems.
- Mapping the Remote Environment to Kyma Environment in which registered APIs and Event catalogs will be used.
- Integrate the registered APIs and Event catalogs with Kyma Service Catalog.

