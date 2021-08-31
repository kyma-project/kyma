---
title: Application Gateway
---

> **CAUTION:** This document describes Application Gateway working in the default legacy [mode](ac-01-application-connector-components.md#application-operator). Read the [details](../ac-02-application-gateway-details.md) to learn about the alternative Compass mode.

Application Gateway sends the requests from Functions and services in Kyma to external APIs registered with Application Registry. Application Gateway works in conjunction with the Access Service, which exposes Application Gateway.

>**NOTE:** The system creates an Access Service for every external API registered by Application Registry.

The following diagram illustrates how Application Gateway interacts with other components and external APIs
which are either unsecured or secured with various security mechanisms and protected against cross-site request forgery (CSRF) attacks.

![Application Gateway Diagram](assets/ac-architecture-proxy-service.svg)

1. A Function calls the Access Service. The name of every Access Service follows this format: `{application-name}-{service-id}`
2. The Access Service exposes Application Gateway.
3. Application Gateway extracts the Application name and the service ID from the name of the Access Service name. Using the extracted Application name, Application Gateway finds the respective Application custom resource and obtains the information about the registered external API, such as the API URL and security credentials.
4. Application Gateway gets a token from the OAuth server.
5. Application Gateway gets a CSRF token from the endpoint exposed by the upstream service. This step is optional and is valid only for the API which was registered with a CSRF token turned on.
6. Application Gateway calls the target API.

## Caching

To ensure optimal performance, Application Gateway caches the OAuth tokens and CSRF tokens it obtains. If the service doesn't find valid tokens for the call it makes, it gets new tokens from the OAuth server and the CSRF token endpoint.
Additionally, the service caches ReverseProxy objects used to proxy the requests to the underlying URL.

## Handling of headers

Application Gateway removes the following headers while making calls to the registered Applications:

- `X-Forwarded-Proto`
- `X-Forwarded-For`
- `X-Forwarded-Host`
- `X-Forwarded-Client-Cert`

In addition, the `User-Agent` header is set to an empty value not specified in the call, which prevents setting the default value.