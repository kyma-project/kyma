---
title: Compass components
type: Architecture
---

The Management Plane is an abstract definition and set of exposed functionality on how users can managed different aspects of their application landscape allowing flexible approaches of extending, customizing and integrating their existing application solutions.

The Management Plane consists of the Management Plane Services (Project Compass), Manage Plane Integration Services, Runtime Provisioners and Cockpit components. The Management Plane Services (Project Compass) are a set of headless services covering all the generic functionality while optionally leveraging different application specific Management Plane Integration Services to configure and instrument the application to be integrated or extended. All communication, whether it comes from a Applications or other external component is flowing through the API-Gateway component. Administrator uses Cockpit to configure Management Plane.


![Management Plane Components](./assets/mp-components.svg)


## Management Plane Integration Services
not oss, musi sobie dopisac zeby moc przylaczac services

## Cockpit

Cockpit component calls Management Plane APIs (in particular Compass and Runtime Provisioner APIs). This component is interchangeable.

## Connector

Connector component exposes GraphQL API that can be accessed directly, its responsibility is establishing trust among Applications, Management Plane and Runtimes.

## API-Gateway

API-Gateway component serves as the main Gateway that extracts Tenant from incoming requests and proxies the requests to the Director component.

## Director

Director component exposes GraphQL API that can be accessed through the Gateway component. It contains all business logic required to handle Applications and Runtimes registration as well as health checks. It also requests Application Webhook API for credentials. This component has access to storage.


## Runtime Provisioner

Runtime Provisioner - zewnetrzny kompoennent handles the creation, modification, and deletion of Runtimes. This component is interchangeable. not oss, separate , stawia kluster
