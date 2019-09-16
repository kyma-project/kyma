---
title: Compass components
type: Architecture
---

Compass consists of a set of components that allow you to extend, customize, and integrate your applications:

![Components](./assets/components.svg)

## Cockpit

Cockpit is a UI that calls Compass APIs. This component is interchangeable.

## API-Gateway

API Gateway serves as the main gateway that proxies the tenant's incoming requests to the Director component. All communication, whether it comes from an application or other external components, flows through API-Gateway.

## Connector

Connector establishes trust between applications, Runtimes, and Compass. Currently, only client certificates are supported.

## Director

Director handles the process of registering applications and Runtimes. It also requests application webhook APIs for credentials and exposes health information about Runtimes. This component has access to the storage.

## Runtime Provisioner

Runtime Provisioner handles the creation, modification, and deletion of Runtimes. This component is interchangeable.

## Central Integration Service

Central Integration Service provides integration with Compass for the whole class of applications. It manages multiple instances of these applications. You can integrate multiple central services to support different types of applications.
