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
- **externalAPIPort** - This port exposes the Event-Service API to an external solution. The default port is `8081`.
- **eventsTargetURL** - A URL to which you proxy the incoming Events. The default URL is `http://localhost:9000`.
- **requestTimeout** - A time-out for requests sent through the Event-Service. It is provided in seconds. The default time-out is `1`.
- **requestLogging** - A flag for logging incoming requests. The default value is `false`.
- **sourceId** - The identifier of the Events' source.

### Unit tests

To run the unit tests, use the following command:

```
go test `go list ./internal/... ./cmd/...`
```

### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
