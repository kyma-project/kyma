---
title: Application Connector components
---

![Architecture Diagram](assets/ac-application-connector-architecture.svg)

## Istio Ingress Gateway

The Istio Ingress Gateway exposes Application Connector and other Kyma components.
The DNS name of the Ingress is cluster-dependent and follows the `gateway.{cluster-dns}` format. For example, `gateway.servicemanager.cluster.kyma.cx`.
Istio Ingress Gateway secures the endpoints with certificate validation. Each call must include a valid client certificate.
You can access every exposed Application using the assigned path. For example, to reach the Gateway for the `user-custom` Application, use `gateway.servicemanager.cluster.kyma.cx/user-custom`.

## Application Connectivity Validator

Application Connectivity Validator verifies the subject of the client certificate, and proxies requests to Application Registry and Event Publisher.

## Connector Service

>**CAUTION:** Connector Service is used only in the [legacy mode](../../01-overview/main-areas/application-connectivity/README.md) of Application Connectivity. 

Connector Service:

- Handles the exchange of client certificates for a given Application.
- Provides the Application Registry and Event Publisher endpoints.
- Signs client certificates using the server-side certificate stored in a Kubernetes Secret.

## Application Registry

>**CAUTION:** Application Registry is used only in the [legacy mode](../../01-overview/main-areas/application-connectivity/README.md) of Application Connectivity.

Application Registry saves and reads the APIs and Event Catalog metadata of the connected external solution in the [Application](../../05-technical-reference/00-custom-resources/ac-01-application.md) custom resource (CR).
The system creates a new Kubernetes service for each registered API.

>**NOTE:** Using Application Registry, you can register an API along with its OAuth or Basic Authentication credentials. The credentials are stored in a Kubernetes Secret.

## Event Publisher

Event Publisher sends events to Eventing with metadata that indicates the source of the event.
This allows routing events to Functions and services based on their source Application.

## Application

An Application represents an external solution connected to Kyma. It handles the integration with other components, such as Eventing.
Using the components of Application Connector, the Application creates a coherent identity for a connected external solution and ensures its separation.
All Applications are instances of the Application custom resource, which also stores all of the relevant metadata.

>**NOTE:** Every Application custom resource corresponds to a single Application to which you can connect an external solution.

## Application Gateway

Application Gateway is an intermediary component between a Function or a service and an external API.
It [proxies the requests](./ac-03-application-gateway.md) from Functions and services in Kyma to external APIs based on the configuration stored in Secrets.

Application Gateway can call services which are not secured, or are secured with:

- [Basic Authentication](https://tools.ietf.org/html/rfc7617)
- OAuth
- Client certificates

Additionally, Application Gateway supports cross-site request forgery (CSRF) tokens as an optional layer of API protection.


## Kubernetes Secret

The Kubernetes Secret is a Kubernetes object which stores sensitive data, such as the OAuth credentials.
