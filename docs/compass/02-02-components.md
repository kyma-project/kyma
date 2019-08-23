---
title: Compass components
type: Architecture
---

Compass consists of a set of components that allow you to extend, customize and integrate your applications.

![Components](./assets/components.svg)

## Cockpit

Cockpit is a UI that calls Compass APIs. This component is interchangeable.

## API-Gateway

API-Gateway serves as the main gateway that proxies tenant's incoming requests to the Director component. All communication, whether it comes from an applications or other external component, flows through API-Gateway.

## Connector

Connector establishes trust between applications, runtimes, and Compass. Currently, only client certificates are supported.

## Director

Director handles the process of applications and runtimes registration. It also requests appliction webhook APIs for credentials and exposes health information about runtimes. This component has access to the storage.

## Runtime Provisioner

Runtime Provisioner handles creation, modification, and deletion of runtimes. This component is interchangeable.

## Central Integration Service

Central Integration Services provides integration with Compass for the whole class of applications. It manages multiple instances of these applications. You can integrate multiple central services to support different types of applications.
