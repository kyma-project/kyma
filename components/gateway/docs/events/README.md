# Event processing application

## Overview
The Event process as part of the Gateway exposes the [Events API](https://github.com/kyma-project/kyma/components/gateway/blob/master/docs/events/api.yaml) to an external application that publishes Events to Kyma.

It uses a mock implementation of the `Receiver End Point`.
This mock dumps an incoming HTTP request and returns it in the body of the response that the `event-id` finds in the request.

## Build
A Makefile is located in the `cmd/events` folder.
- To build the event processing application, use this command:
```
$make compile
```
This command compiles and runs the unit tests.
The event process is available in the `cmd/events/bin` folder.
- To build a `docker image`, use this command:
```
make docker-build`
```

## Command line arguments
These are the available command line arguments:

```
events -help
  -dump_requests
        Dump the incoming request
  -help
        Print the command line options
  -max_requests int
        The max number of accepted concurrent requests, 0 means no limitation
  -port int
        The events/publish listen port (default 8080)
  -source_environment string
        The ID of the event source (default "stage")
  -source_namespace string
        The organization publishing the event
  -source_type string
        The type of the event source
```

## Test the Event API
Use the SwaggerUI plugin of your IDE to load and test the [Events API](https://github.com/kyma-project/kyma/blob/master/components/gateway/docs/api/externalapi.yaml#L166), or do it directly with this command:
```
curl -v -X POST "localhost:8080/v1/events" -H "accept: application/json" -H "Content-Type: application/json" -d "{\"event-type\":\"order.created\",\"event-type-version\":\"v1\",\"event-id\":\"31109198-4d69-4ae0-972d-76117f3748c8\",\"event-time\":\"2012-11-01T22:08:41+00:00\", \"data\":\"my order created\" }"
```
