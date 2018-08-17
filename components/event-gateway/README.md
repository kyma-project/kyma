# Event-Gateway

## Overview

This is the repository for the Kyma Event-Gateway.

## Installation

To install the Event-Gateway, follow these steps:

1. `git clone git@github.com:kyma-project/kyma.git`
1. `cd kyma/components/event-gateway`
1. `CGO_ENABLED=0 go build ./cmd/eventgateway`

## Usage

The Event-Gateway has the following parameters:
- **externalAPIPort** - This port exposes the Event-Gateway API to an external solution. The default port is `8081`.
- **eventsTargetURL** - A URL to which you proxy the incoming Events. The default URL is `http://localhost:9000`.
- **requestTimeout** - A time-out for requests sent through the Event-Gateway. It is provided in seconds. The default time-out is `1`.
- **requestLogging** - A flag for logging incoming requests. The default value is `false`.

The parameters for the Event API correspond to the fields in the [Remote Environment](https://github.com/kyma-project/kyma/tree/master/docs/remote-environment.md):

- **sourceEnvironment** - The name of the Event source environment.
- **sourceType** - The type of the Event source.
- **sourceNamespace** - The organization that publishes the Event.

### Unit tests

To run the unit tests, use the following command:

```
go test `go list ./internal/... ./cmd/...`
```

### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
