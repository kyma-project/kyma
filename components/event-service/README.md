# Event-Service

## Overview

This is the repository for the Kyma Event-Service.

## Installation

To install the Event-Service, follow these steps:

1. `git clone git@github.com:kyma-project/kyma.git`
1. `cd kyma/components/event-service`
1. `CGO_ENABLED=0 go build ./cmd/eventservice`

## Usage

The Event-Service has the following parameters:
- **externalAPIPort** is the port that exposes the Event Service API to an external solution. The default port is `8081`.
- **eventsTargetURL** is the URL to which you proxy the incoming Events. The default URL is `http://localhost:9000`.
- **maxRequestSize** is the maximum publish request body size in bytes. The default is `65536`.
- **requestTimeout** is the timeout for requests sent through the Event Service. It is provided in seconds. The default value is `1`.
- **requestLogging** is the flag for logging incoming requests. The default value is `false`.
- **sourceId** is the identifier of the Events' source.

### Unit tests

To run the unit tests, use the following command:

```
go test `go list ./internal/... ./cmd/...`
```

### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
