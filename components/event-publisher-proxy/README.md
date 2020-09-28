# Event Publisher Proxy

## Overview

The Event Publisher Proxy receives Cloud Event publishing requests from the cluster workloads (microservice or Serverless functions) and redirects them to the Enterprise Messaging Service Cloud Event Gateway.

## Prerequisites

- Go modules
- [ko](https://github.com/google/ko)

## Development

### Build

```bash
$ go mod vendor
```

### Test

```bash
$ make test-local
```

### Deploy inside a cluster

```bash
$ ko apply -f config/
```

### Send Events

```bash
curl -v -X POST \
    -H "Content-Type: application/cloudevents+json" \
    --data @<(<<EOF
    {
        "specversion": "1.0",
        "source": "/default/sap.kyma/kt1",
        "type": "sap.kyma.FreightOrder.Arrived.v1",
        "eventtypeversion": "v1",
        "id": "A234-1234-1234",
        "data" : "{\"foo\":\"bar\"}",
        "datacontenttype":"application/json"
    }
EOF
    ) \
    http://<hostname>/publish
```

## Environment Variables

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
