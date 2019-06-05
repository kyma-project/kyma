---
title: Overview
---

The Application Connector (AC) is a proprietary Kyma implementation that allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process doesn't change, which allows to avoid any configuration or network-related problems.

The external solution you connect to Kyma using the AC is represented as an Application (App). There is always a one-to-one relationship between a connected solution and an App, which helps to ensure the highest level of security and separation. This means that you must create five separate Apps in your cluster to connect five different external solutions and use their APIs and Event catalogs in Kyma.

The AC gives you this functionality:

- Manages the lifecycle of Apps.
- Establishes a secure connection and generates the client certificate used by the connected external solution.
- Registers the APIs and the Event catalogs of the connected external solution.
- Delivers the Events from the connected external solution to the Kyma Event Bus.
- Proxies calls sent from Kyma to external APIs registered by the connected external solution.
- Allows to map an App to a Kyma Namespace and use its registered APIs and Event catalogs in the context of that Namespace.
- Integrates the registered APIs and Event catalogs with the Kyma Service Catalog.

All of the AC components scale independently, which allows to adjust it to fit the needs of the implementation built using Kyma.


## Supported APIs

The Application Connector allows you to register secured REST APIs exposed by the connected external solution. The Application Connector supports a variety of authentication methods to ensure smooth integration with a wide range of APIs. 

You can register an API secured with one of the following authentication methods:

- Basic Authentication
- OAuth
- Client Certificates

> **NOTE:** You can register non-secured APIs for testing purposes, however, it is not recommended in the production environment.

In addition to authentication methods, the Application Connector supports Cross-Site Request Forgery Tokens.

You can register any API that adheres to the REST principles and is available over the HTTP protocol. The Application Connector also allows you to register APIs implemented with the OData technology. 

You can provide specifications that document your APIs. The Application Connector supports [OpenAPI](https://www.openapis.org/) and [OData](https://www.odata.org/documentation) specification formats.