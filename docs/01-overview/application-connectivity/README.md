---
title: What is Application Connectivity in Kyma?
---

Application Connectivity in Kyma is an area that: 

- Simplifies and secures the connection between external systems and Kyma
- Stores and handles the metadata of external systems
- Provides certificate handling for the [Eventing](../eventing/README.md) flow in the Compass scenario (mode)
- Manages secure access to external systems
- Provides monitoring and tracing capabilities to facilitate operational aspects

Depending on your use case, Application Connectivity works in one of the two modes: 
- **Standalone mode** (default) - a standalone mode where Kyma is not connected to [Compass](https://github.com/kyma-incubator/compass)
- **Compass mode** - using [Runtime Agent](ra-01-runtime-agent-overview.md) and integration with [Compass](https://github.com/kyma-incubator/compass) to automate connection and registration of services using mTLS certificates

# Application Connector

## Overview

Application Connector (AC) is a custom, in-house built Kyma component that allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process does not change, which allows you to avoid any configuration or network-related problems.

The external solution you connect to Kyma using AC is represented as an Application. There is always a one-to-one relationship between a connected solution and an Application, which helps to ensure the highest level of security and separation. This means that you must create five separate Applications in your cluster to connect five different external solutions and use their APIs and event catalogs in Kyma.

Application Connector secures Eventing with a client certificate verified by the Istio Ingress Gateway in the [Compass scenario](./README.md).

>**NOTE:** When using AC, make sure to [enable automatic Istio sidecar proxy injection](/istio/user/02-operation-guides/operations/02-20-enable-sidecar-injection). For more details, see [Default Istio setup in Kyma](/istio/user/00-overview/00-40-overview-istio-setup).

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

# Application Gateway

Application Gateway is an intermediary component between a Function or a microservice and an external API. 
It [proxies the requests](../../05-technical-reference/00-architecture/ac-03-application-gateway.md) from Functions and microservices in Kyma to external APIs based on the configuration stored in Secrets.

Application Gateway also supports redirects for the request flows in which the URL host remains unchanged. For more details, see [Response rewriting](../../05-technical-reference/ac-01-application-gateway-details.md#response-rewriting).

## Supported API authentication for Application CR

Application Gateway can call services which are not secured, or are secured with:

- [Basic Authentication](https://tools.ietf.org/html/rfc7617)
- [OAuth](https://tools.ietf.org/html/rfc6750)
- [OAuth 2.0 mTLS](https://datatracker.ietf.org/doc/html/rfc8705)
- Client certificates

Additionally, Application Gateway supports cross-site request forgery (CSRF) tokens as an optional layer of API protection.

Application Gateway calls the registered APIs accordingly, basing on the security type specified for the API in the Application CR.

## Provide a custom access token

Application Gateway overrides the registered API's security type if it gets a request which contains the **Access-Token** header. In such a case, Application Gateway rewrites the token from the **Access-Token** header into an OAuth-compliant **Authorization** header and forwards it to the target API.

This mechanism is suited for implementations in which an external application handles user authentication.

See how to [pass an access token in a request header](../../04-operation-guides/operations/ac-01-pass-access-token-in-request-header.md).

# Security

## Client certificates

To provide maximum security, in the [Compass mode](./README.md), Application Connector uses the TLS protocol with Client Authentication enabled. As a result, whoever wants to connect to Application Connector must present a valid client certificate, which is dedicated to a specific Application. In this way, the traffic is fully encrypted and the client has a valid identity.

### TLS certificate verification for external systems

By default, the TLS certificate verification is enabled when sending data and requests to every application.
You can [disable the TLS certificate verification](../../03-tutorials/00-application-connectivity/ac-11-disable-tls-certificate-verification.md) in the communication between Kyma and an application to allow Kyma to send requests and data to an unsecured application. Disabling the certificate verification can be useful in certain testing scenarios.

# Runtime agent

Runtime Agent is a Kyma component that connects to [Compass](https://github.com/kyma-incubator/compass). It is an integral part of every Kyma Runtime in the [Compass mode](README.md) and it fetches the latest configuration from Compass. It also provides Runtime-specific information that is displayed in the Compass UI, such as Runtime UI URL, and it provides Compass with Runtime configuration, such as Event Gateway URL, that should be passed to an Application. To learn more, read the section on [configuring the Runtime](../../05-technical-reference/ra-01-configuring-runtime.md).

The main responsibilities of the component are:
- Establishing a trusted connection between the Kyma Runtime and Compass
- Renewing a trusted connection between the Kyma Runtime and Compass
- Synchronizing with the [Director](https://github.com/kyma-incubator/compass/blob/master/docs/compass/02-01-components.md#director) by fetching new Applications from the Director and creating them in the Runtime, and removing from the Runtime Applications that no longer exist in the Director.

# Useful links

If you're interested in learning more about the Application Connectivity area, follow these links to:

- Perform some simple and more advanced tasks:

  - [Pass the access token in the request header](../../04-operation-guides/operations/ac-01-pass-access-token-in-request-header.md)
  - [Create a new Application](../../03-tutorials/00-application-connectivity/ac-01-create-application.md)
  - [Register a service](../../03-tutorials/00-application-connectivity/ac-03-register-manage-services.md)
  - [Register a secured API](../../03-tutorials/00-application-connectivity/ac-04-register-secured-api.md)
  - [Call a registered external service from Kyma](../../03-tutorials/00-application-connectivity/ac-05-call-registered-service-from-kyma.md)
  - [Disable TLS certificate verification](../../03-tutorials/00-application-connectivity/ac-11-disable-tls-certificate-verification.md)
  - [Enable Kyma with Runtime Agent](../../04-operation-guides/operations/ra-01-enable-kyma-with-runtime-agent.md)
  - [Establish a secure connection with Compass](../../03-tutorials/00-application-connectivity/ra-01-establish-secure-connection-with-compass.md)
  - [Maintain a secure connection with Compass](../../03-tutorials/00-application-connectivity/ra-02-maintain-secure-connection-with-compass.md)
  - [Revoke a client certificate (RA)](../../03-tutorials/00-application-connectivity/ra-03-revoke-client-certificate.md)
  - [Configure Runtime Agent with Compass](../../03-tutorials/00-application-connectivity/ra-04-configure-runtime-agent-with-compass.md)
  - [Reconnect Runtime Agent with Compass](../../03-tutorials/00-application-connectivity/ra-05-reconnect-runtime-agent-with-compass.md)

- Analyze Application Connectivity specification and configuration files:

  - [Application](../../05-technical-reference/00-custom-resources/ac-01-application.md) custom resource (CR)
  - [Application Connector chart](../../05-technical-reference/00-configuration-parameters/ac-01-application-connector-chart.md)
  - [Compass Connection](../../05-technical-reference/00-custom-resources/ra-01-compassconnection.md) custom resource (CR)
  - [Connection with Compass](../../05-technical-reference/00-configuration-parameters/ra-01-connection-with-compass.md) 

- Understand technicalities behind the Application Connectivity implementation:

  - [Application Connector components](../../05-technical-reference/00-architecture/ac-01-application-connector-components.md)
  - [Application Gateway workflow](../../05-technical-reference/00-architecture/ac-03-application-gateway.md)
  - [Application Gateway details](../../05-technical-reference/ac-01-application-gateway-details.md)
  - [Runtime Agent workflow](../../05-technical-reference/00-architecture/ra-01-runtime-agent-workflow.md)
  - [Configuring the Runtime](../../05-technical-reference/ra-01-configuring-runtime.md)
