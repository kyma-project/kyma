---
title: Overview
---

The Application Connector is a proprietary Kyma implementation that allows you to connect with external solutions. The Application Connector consists of four
components that ensure the security of the connection and the access to all of the external solution's Events and APIs. The implementation handles routing of the calls and Events coming from an external solution to Kyma, and the API calls sent from Kyma to the connected external solution.

These are the components of the Application Connector:

- The **Connector Service** generates the required certificates and ensures a secure and trusted connection between Kyma and an external solution. This is a global service that works in the context of
a given Remote Environment.
- The **Metadata Service** allows you to register all of the external solution's APIs and Event catalogs which Kyma consumes. You can register the APIs along with additional documentation and Swagger files.
This is a global service that works in the context of a given Remote Environment.
- The **Gateway Service** proxies the API calls sent from Kyma to the connected external solution and handles OAuth2 tokens. A new instance of this service is deployed for every Remote Environment.
- The **Event Service** delivers the Events sent from a connected external solution to Kyma. A new instance of this service is deployed for every Remote Environment.

To ensure maximum security and separation, a single Remote Environment allows you to connect only to a single external solution.
