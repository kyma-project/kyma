---
title: The Application Connector
type: Overview
---

The Application Connector is a proprietary Kyma implementation that allows you to connect with external solutions. The Application Connector consists of three
components that ensure the security of the connection, the access to all of the external solution's Events and APIs, as well as proper routing of the calls and Events coming from an external solution to Kyma, and the API calls sent from Kyma to the connected external solution.

- The **Connector Service** generates the required certificates and ensures a secure and trusted connection between Kyma and an external solution.
- The **Metadata Service** allows you to register all of the external solution's APIs and Event catalogs which Kyma consumes.
- The **Gateway Service** proxies the calls and Events sent between Kyma and the registered APIs an external solution, and routes the traffic to the appropriate components and APIs on both sides of the connection.

To ensure maximum security and separation, a single instance of the Gateway Service allows you to connect only to a single external solution. This connection is represented in Kyma by a Remote Environment.
