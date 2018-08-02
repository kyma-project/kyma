---
title: The Application Connector
type: Overview
---

The Application Connector is a proprietary Kyma implementation that allows you to connect with external solutions. The Application Connector consists of three
components that ensure the security of the connection, and the access to all of the external solution's Events and APIs. The implementation handles routing of the calls and Events coming from an external solution to Kyma, and the API calls sent from Kyma to the connected external solution.

These are the components of the Application Connector:

- The **Connector Service** generates the required certificates and ensures a secure and trusted connection between Kyma and an external solution.
- The **Metadata Service** allows you to register all of the external solution's APIs and Event catalogs which Kyma consumes. You can register the APIs along with additional documentation and Swagger files.
- The **Gateway Service** proxies the API calls sent from Kyma to the connected external solution and handles OAuth2 tokens. 
- The **Event Service** delivers the Events sent from a connected external solution to Kyma.

To ensure maximum security and separation, a single instance of the Gateway Service allows you to connect only to a single external solution. This connection is represented in Kyma by a [Remote Environment](./014-details-remote-environment.md).
