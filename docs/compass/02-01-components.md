---
title: Components
type: Architecture
---

![Components](./assets/components.svg)

## Application

Application represents any external system that you want to register to Compass with its API and Event definitions. These are the types of possible integration levels between an application and Compass:
- Manual integration - the administrator manually provides API or Events metadata to Compass. This type of integration is used mainly for simple use-case scenarios and doesn't support all features.
- Built-in integration - integration with Compass is built in the application.
- Proxy - a highly application-specific proxy component provides the integration.
- Central integration service -  a central service provides integration for the dedicated group of applications. It manages multiple instances of these applications. You can integrate multiple central services to support different types of applications.

## Kyma Runtime

Runtime is any system to which you can apply configuration provided by Compass. Your Runtime must get a trusted connection to Compass. It must also allow for fetching application definitions and using these applications in a given tenant.

By default, Compass is integrated with Kyma (Kubernetes), but its usage can also be extended to other platforms, such as CloudFoundry or Serverless.

## Agent

Agent is an integral part of every Kyma Runtime, which fetches the latest configuration from Compass. In the future releases, Agent will also send information about Runtime health checks to Compass.

## Cockpit

Cockpit is a UI that calls Compass APIs. This component is interchangeable.

## Gateway

Gateway proxies the tenant's incoming requests to the Director component. All communication, whether it comes from an application or other external components, flows through Gateway.

## Connector

Connector establishes trust between applications and Runtimes. Currently, only client certificates are supported.

## Director

Director handles the process of registering applications and Runtimes. It also requests application webhook APIs for credentials and exposes health information about Runtimes. This component has access to the storage.

## Runtime Provisioner

Runtime Provisioner handles the creation, modification, and deletion of Runtimes. This component is interchangeable.

## Central Integration Service

Central Integration Service provides integration with Compass for the whole class of applications. It manages multiple instances of these applications. You can integrate multiple central services to support different types of applications.
