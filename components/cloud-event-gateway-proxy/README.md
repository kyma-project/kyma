# Cloud Event Gateway Proxy

## Overview

The Cloud Event Gateway proxy receives Cloud Event publishing requests from the cluster workloads (microservice or serverless functions) and redirects it to the Enterprise Messaging Service Cloud Event Gateway.

## Prerequisites

- go modules.
- [ko](https://github.com/google/ko).

## Development

Build

```bash
$ go mod vendor
```

Test

```bash
$ make test-local
```

Deploy inside a cluster

```bash
$ ko apply -f config/
```

## Environment Variables

| Environment Variable    | Description                                                                                   | Default Value |
| ----------------------- | --------------------------------------------------------------------------------------------- | ------------- |
| INGRESS_PORT            | The ingress port for the CloudEvents Gateway Proxy.                                           | 8080          |
| MAX_IDLE_CONNS          | The maximum number of idle (keep-alive) connections across all hosts. Zero means no limit.    | 100           |
| MAX_IDLE_CONNS_PER_HOST | The maximum idle (keep-alive) connections to keep per-host. Zero means use the default value. | 2             |
| CLIENT_ID               | The Client ID used to acquire Access Tokens from the Authentication server.                   |               |
| CLIENT_SECRET           | The Client Secret used to acquire Access Tokens from the Authentication server.               |               |
| TOKEN_ENDPOINT          | The Authentication Server Endpoint to provide Access Tokens.                                  |               |
| EMS_CE_URL              | The Messaging Server Endpoint that accepts publishing CloudEvents to it.                      |               |

