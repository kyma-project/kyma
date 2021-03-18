# Event Publisher Proxy

## Overview

The Event Publisher Proxy receives legacy and Cloud Event publishing requests from the cluster workloads (microservice or Serverless functions) and redirects them to the Enterprise Messaging Service Cloud Event Gateway. It also fetches a list of subscriptions for a connected application.

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
$ ko apply -f config/event-publisher-proxy/
```

### Send Events

This command supports **cloud events**: 
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

This command supports **legacy events**:
```bash
curl -v -X POST \
    -H "Content-Type: application/json" \
    --data @<(<<EOF
    {
        "event-type": "order.created",
        "event-type-version": "v0",
        "event-time": "2020-04-02T21:37:00Z",
        "data" : "{\"foo\":\"legacy-mode-on\"}"
    }
EOF
    ) \
    http://<hostname>/application-name/v1/events
```

### Get a list of subscriptions for a connected application

```bash
curl -v -X GET \
    -H "Content-Type: application/json" \
    http://hostname/:application-name/v1/events/subscribed
```

## Environment Variables

| Environment Variable    | Default Value | Description                                                                                |
| ----------------------- | ------------- |------------------------------------------------------------------------------------------- |
| INGRESS_PORT            | 8080          | The ingress port for the CloudEvents Gateway Proxy.                                        |
| MAX_IDLE_CONNS          | 100           | The maximum number of idle (keep-alive) connections across all hosts. Zero means no limit. |
| MAX_IDLE_CONNS_PER_HOST | 2             | The maximum idle (keep-alive) connections to keep per-host. Zero means the default value.  |
| REQUEST_TIMEOUT         | 5s            | The timeout for the outgoing requests to the Messaging server.                             |
| CLIENT_ID               |               | The Client ID used to acquire Access Tokens from the Authentication server.                |
| CLIENT_SECRET           |               | The Client Secret used to acquire Access Tokens from the Authentication server.            |
| TOKEN_ENDPOINT          |               | The Authentication Server Endpoint to provide Access Tokens.                               |
| EMS_PUBLISH_URL         |               | The Messaging Server Endpoint that accepts publishing CloudEvents to it.                   |
| BEB_NAMESPACE           |               | The name of the namespace in BEB.                                                          |
| EVENT_TYPE_PREFIX       |               | The prefix of the eventType as per the BEB event specification.                            |

## Flags
| Flag                    | Default Value | Description                                                                                |
| ----------------------- | ------------- |------------------------------------------------------------------------------------------- |
| max-request-Size        | 65536         | The maximum size of the request.                                                           |
| metrics-addr            | :9090         | The address the metric endpoint binds to.                                                  |
