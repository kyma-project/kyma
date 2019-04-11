---
title: APIs exposed by the external solution
type: Overview
---

The Application Connector allows you to register secured REST APIs exposed by the connected external solution. The Application Connector supports a variety of authentication methods to ensure smooth integration with a wide range of APIs. 

You can register an API secured with one of the following authentication methods:

- Basic Authentication
- OAuth
- Client Certificates

> **NOTE:** You can register non-secured APIs for testing purposes, however, it is not recommended in the production environment.

In addition to authentication methods, the Application Connector supports Cross-Site Request Forgery Tokens.

You can register any API that adheres to the REST principles and is available over the HTTP protocol. The Application Connector also allows you to register APIs implemented with the OData technology. 

You can provide specifications that document your APIs. The Application Connector supports [OpenAPI](https://www.openapis.org/) and [OData](https://www.odata.org/documentation) specification formats.
