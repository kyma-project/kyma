---
title: Overview
---

The Application Connector (AC) is a proprietary Kyma implementation that allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process does not change, which allows to avoid any configuration or network-related problems.

The external solution you connect to Kyma using the AC is represented as an Application. There is always a one-to-one relationship between a connected solution and an Application, which helps to ensure the highest level of security and separation. This means that you must create five separate Applications in your cluster to connect five different external solutions and use their APIs and event catalogs in Kyma.

The Application Connector:

- Manages lifecycles of Applications.
- Establishes a secure connection and generates the client certificate used by the connected external solution.
- Registers APIs and event catalogs of the connected external solution.
- Delivers events from the connected external solution to the Kyma Event Bus.
- Proxies calls sent from Kyma to external APIs registered by the connected external solution.
- Allows to map an Application to a Kyma Namespace and use its registered APIs and event catalogs in the context of that Namespace.
- Integrates the registered APIs and event catalogs with the Kyma Service Catalog.

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

You can provide specifications that document your APIs. The Application Connector supports [AsyncAPI](https://www.asyncapi.com/), [OpenAPI](https://www.openapis.org/), and [OData](https://www.odata.org/documentation) specification formats.

>**TIP:** Follow [this](/components/rafter/#details-asyncapi-service) link to read about the AsyncAPI Service used in Kyma to process AsyncAPI specifications.
