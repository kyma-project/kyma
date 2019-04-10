---
title: APIs exposed by the external solution

---

## Security

The Application Connector allows to register secured REST APIs exposed by the connected external solution. The Application Connector supports a variety of authentication methods to ensure smooth integration with a wide range of APIs. 

Your can register an API secured with one of the following:

- Basic Authentication
- OAuth
- Client Certificates

> **NOTE:** You can register non-secured API for testing purposes, however, it is not recommended in production environment.

In addition to authentication methods the Application Connector supports Cross-Site Request Forgery Tokens.

## Types of APIs supported

You can register any API adhering to the REST principles and available over HTTP protocol. The Application Connector also allows to register APIs implemented with OData technology. 

You can provide specifications documenting your APIs. The Application Connector supports [OpenAPI](https://www.openapis.org/) and [OData](https://www.odata.org/documentation) specification formats.
