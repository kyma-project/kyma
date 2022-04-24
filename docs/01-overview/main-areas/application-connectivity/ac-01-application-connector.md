---
title: Application Connector
---

## Overview

Application Connector (AC) is a custom, in-house built Kyma component that allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process does not change, which allows to avoid any configuration or network-related problems.

The external solution you connect to Kyma using AC is represented as an Application. There is always a one-to-one relationship between a connected solution and an Application, which helps to ensure the highest level of security and separation. This means that you must create five separate Applications in your cluster to connect five different external solutions and use their APIs and event catalogs in Kyma.

Application Connector is secured with a client certificate verified by the Istio Ingress Gateway in the [Compass scenario](./README.md). <!-- TODO: verify -->

## Features

Application Connector:

- Simplifies and secures the connection between external systems and Kyma
- Stores and handles the metadata of external events and APIs
- Proxies calls sent from Kyma to external APIs registered by the connected external solution <!-- TODO: verify -->
- Provides certificate handling for the [Eventing](../eventing/README.md) flow in the [Compass scenario](./README.md)
- Delivers events from the connected external solution to Eventing in the [Compass scenario](./README.md)  <!-- TODO: verify -->
- Manages secure access to external systems 
- Provides monitoring and tracing capabilities to facilitate operational aspects

All the AC components scale independently, which allows to adjust it to fit the needs of the implementation built using Kyma.

## Supported APIs

Application Connector supports secured REST APIs exposed by the connected external solution. Application Connector supports a variety of authentication methods to ensure smooth integration with a wide range of APIs.

The following authentication methods for your secured APIs are supported:

- Basic Authentication
- OAuth
- Client Certificates

> **NOTE:** Non-secured APIs are supported too, however, they are not recommended in the production environment.

In addition to authentication methods, Application Connector supports Cross-Site Request Forgery (CSRF) Tokens.

AC supports any API that adheres to the REST principles and is available over the HTTP protocol. APIs implemented with the OData technology are also supported.

You can provide specifications that document your APIs. Application Connector supports [AsyncAPI](https://www.asyncapi.com/), [OpenAPI](https://www.openapis.org/), and [OData](https://www.odata.org/documentation) specification formats.  <!-- TODO: verify -->

>**TIP:** Read about the [AsyncAPI Service used in Kyma](https://github.com/kyma-project/rafter/blob/main/docs/12-asyncapi-service.md) to process AsyncAPI specifications.  <!-- TODO: verify -->
