---
title: Application Proxy
type: Architecture
---

The Application Proxy sends the requests from Lambda functions and services in Kyma to external APIs registered with the Application Registry. The Application Proxy works in conjunction with the Access Service, which exposes the Application Proxy.

>**NOTE:** The system creates an Access Service for every external API registered by the Application Registry.

The following diagram illustrates how the Application Proxy interacts with other components and external APIs secured with OAuth.

![Proxy Service Diagram](./assets/003-architecture-proxy-service.svg)

1. A lambda function calls the Access Service. The name of every Access Service follows this format: `app-{application-name}-{service-id}`
2. The Access Service exposes the Application Proxy.
3. The Application Proxy extracts the Application name and the service ID from the name of the Access Service name. Using the extracted Application name, the Application Proxy finds the respective Application custom resource and obtains the information about the registered external API, such as the API URL and the OAuth server URL.
4. The Application Proxy gets a token from the OAuth server.
5. The Application Proxy calls the target API using the OAuth token.  

## Caching

To ensure optimal performance, the Application Proxy caches the OAuth tokens it obtains. If the service doesn't find a token valid for the call it makes, it gets a new token from the OAuth server.
Additionally, the service caches ReverseProxy objects used to proxy the requests to the underlying URL.