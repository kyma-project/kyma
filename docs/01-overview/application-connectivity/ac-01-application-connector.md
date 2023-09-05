---
title: Application Connector
---

## Overview

Application Connector (AC) is a custom, in-house built Kyma component that allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process does not change, which allows you to avoid any configuration or network-related problems.

The external solution you connect to Kyma using AC is represented as an Application. There is always a one-to-one relationship between a connected solution and an Application, which helps to ensure the highest level of security and separation. This means that you must create five separate Applications in your cluster to connect five different external solutions and use their APIs and event catalogs in Kyma.

Application Connector secures Eventing with a client certificate verified by the Istio Ingress Gateway in the [Compass scenario](./README.md).

>**NOTE:** When using AC, make sure to [enable automatic Istio sidecar proxy injection](https://kyma-project.io/#/istio/user/02-operation-guides/operations/02-20-enable-sidecar-injection). For more details, see [Default Istio setup in Kyma](https://kyma-project.io/#/istio/user/00-overview/00-40-overview-istio-setup).

## Features

Application Connector:

- Simplifies and secures the connection between external systems and Kyma
- Stores and handles the metadata of external APIs
- Proxies calls sent from Kyma to external APIs registered by the connected external solution 
- Provides certificate handling for the [Eventing](../eventing/README.md) flow in the [Compass scenario](./README.md)
- Delivers events from the connected external solution to Eventing in the [Compass scenario](./README.md) 
- Manages secure access to external systems

All the AC components scale independently, which allows you to adjust it to fit the needs of the implementation built using Kyma.

## Supported APIs

Application Connector supports secured REST APIs exposed by the connected external solution. Application Connector supports a variety of authentication methods to ensure smooth integration with a wide range of APIs.

The following authentication methods for your secured APIs are supported:

- Basic Authentication
- OAuth
- OAuth 2.0 mTLS
- Client Certificates

> **NOTE:** Non-secured APIs are supported too, however, they are not recommended in the production environment.

In addition to authentication methods, Application Connector supports Cross-Site Request Forgery (CSRF) Tokens.

AC supports any API that adheres to the REST principles and is available over the HTTP protocol.
