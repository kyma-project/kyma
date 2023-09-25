---
title: Application Gateway details
---

Application Gateway is an intermediary component between a Function or a microservice and an external API.

## Application Gateway URL

To call a remote system's API from a workload with Application Gateway, you use the URL to the `central-application-gateway.kyma-system` service at an appropriate port and with a respective suffix to access the API of a specific application.

The suffix and the port number differ depending on whether you're using Kyma in the [Standalone or Compass mode](../01-overview/application-connectivity/README.md):

| **Kyma mode** | **Application Gateway URL** |
|-----------|-------------------------|
| Standalone | `http://central-application-gateway.kyma-system:8080/{APP_NAME}/{SERVICE_NAME}/{TARGET_PATH}` |
| Compass | `http://central-application-gateway.kyma-system:8082/{APP_NAME}/{SERVICE_NAME}/{API_ENTRY_NAME}/{TARGET_PATH}` |

The placeholders in the URLs map to the following:

- `APP_NAME` is the name of the Application CR.
- `SERVICE_NAME` represents the API Definition.
- `TARGET_PATH` is the destination API URL.

## Proxying requests

Application Gateway proxies requests from Functions and services in Kyma to external APIs based on the configuration stored in the [Application CR](00-custom-resources/ac-01-application.md) and Kubernetes Secrets.

For examples of configurations and Secrets, see the [tutorial on registering a secured API](../03-tutorials/00-application-connectivity/ac-04-register-secured-api.md).

> **NOTE:** All APIs defined in a single Secret use the same configuration - the same credentials, CSRF tokens, and request parameters.

## Caching

To ensure optimal performance, Application Gateway caches the OAuth tokens and CSRF tokens it obtains. If the service doesn't find valid tokens for the call it makes, it gets new tokens from the OAuth server and the CSRF token endpoint.
Additionally, the service caches ReverseProxy objects used to proxy requests to the underlying URL.

## Handling of headers

Application Gateway proxies the following headers while making calls to the registered Applications:

- `X-Forwarded-Proto`
- `X-Forwarded-For`
- `X-Forwarded-Host`
- `X-Forwarded-Client-Cert`

In addition, the `User-Agent` header is set to an empty value not specified in the call, which prevents setting the default value.

## Response rewriting

Application Gateway performs response rewriting in situations when during a call to the external system, the target responds with a redirect (`3xx` status code) that points to the URL with the same host and a different path.
In such a case, the `Location` header is modified so that the original target path is replaced with the Application Gateway URL and port. The sub-path pointing to the called service remains attached at the end. 
The modified `Location` header has the following format: `{APP_GATEWAY_URL}:{APP_GATEWAY_PORT}/{APP_NAME}/{SERVICE_NAME}/{SUB-PATH}`.

This functionality makes the HTTP clients that originally called Application Gateway follow redirects through the Gateway, and not to the service directly. 
This allows for passing authorization, custom headers, URL parameters, and the body without an issue.

Application Gateway also rewrites all the `5xx` status codes to a `502`. In such a case, the `Target-System-Status` header contains the original code returned by the target. 
