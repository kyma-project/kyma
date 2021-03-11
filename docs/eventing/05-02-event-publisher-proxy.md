---
title: Event Publisher Proxy configuration
type: Configuration
---

The Event Publisher Proxy receives legacy and Cloud Event publishing requests from the cluster workloads (microservice or Serverless functions) and redirects them to the Enterprise Messaging Service Cloud Event Gateway. It also fetches a list of subscriptions for a connected application.

## Environment variables

This table shows the environment variables that are used by the Event Publisher Proxy.

| Environment Variable    | Default Value | Description                                                                                   |
| ----------------------- | ------------- |---------------------------------------------------------------------------------------------- |
| INGRESS_PORT            | 8080          | The ingress port for the CloudEvents Gateway Proxy.                                           |
| MAX_IDLE_CONNS          | 100           | The maximum number of idle (keep-alive) connections across all hosts. Zero means no limit.    |
| MAX_IDLE_CONNS_PER_HOST | 2             | The maximum idle (keep-alive) connections to keep per-host. Zero means the default value.     |
| REQUEST_TIMEOUT         | 5s            | The timeout for the outgoing requests to the Messaging server.                                |
| CLIENT_ID               |               | The Client ID used to acquire Access Tokens from the Authentication server.                   |
| CLIENT_SECRET           |               | The Client Secret used to acquire Access Tokens from the Authentication server.               |
| TOKEN_ENDPOINT          |               | The Authentication Server Endpoint to provide Access Tokens.                                  |
| EMS_PUBLISH_URL         |               | The Messaging Server Endpoint that accepts publishing CloudEvents to it.                      |
| BEB_NAMESPACE           |               | The name of the namespace in BEB.                                                        |
| EVENT_TYPE_PREFIX       |               | The prefix of the eventType as per the BEB event specification.                                    |


## Flags

This table shows flags used by the Event Publisher Proxy.

| Flag | Default Value | Description                                                                                   |
| ----------------------- | ------------- |---------------------------------------------------------------------------------------------- |
| maxRequestSize | 65536 | The maximum size of the request. |