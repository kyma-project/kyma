---
title: Compass components
type: Architecture
---

Compass consists of a set of components that allow you to extend, customize and integrate your applications. Administrator uses Cockpit to configure Compass.

![Components](./assets/components.svg)

## Cockpit

Cockpit is a UI that calls Compass APIs. This component is interchangeable.

## API-Gateway

API-Gateway serves as the main gateway that proxies tenant's incoming requests to the Director component. All communication, whether it comes from an applications or other external component, flows through API-Gateway.

## Connector

Connector establishes trust between applications, runtimes, and Compass.

## Director

Director exposes GraphQL API that can be accessed through the Gateway component. It contains all business logic required to handle applications and runtimes registration as well as health checks. It also requests application webhook API for credentials. This component has access to storage. It contains the Registry component that serves as the persistent storage interface.

## Runtime Provisioner

Runtime Provisioner handles creation, modification, and deletion of runtimes. This component is interchangeable.

## Central Integration Services

Central Integration Services provides integration with Compass for the whole class of applications. It manages multiple instances of these applications. You can integrate multiple central services to support different types of applications.
