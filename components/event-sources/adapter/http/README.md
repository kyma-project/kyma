# HTTP Source Adapter

## Overview

The HTTP Source adapter is an HTTP server that receives [CloudEvents](https://github.com/cloudevents/spec) in version 1.0 and proxies them to a preconfigured sink.
The adapter is written using [Go SDK for CloudEvents](https://github.com/cloudevents/sdk-go) and is fully compatible with the CloudEvents 1.0 specification.


It accepts binary and structured content modes. The batched mode is not implemented.
See [this document](https://github.com/cloudevents/spec/blob/master/http-protocol-binding.md#13-content-modes) for details about the content types and the HTTP protocol binding.

## Usage

### Run the adapter locally

To use the adapter locally:

1. Start with exporting the variables. Run:

```bash
export SINK_URI=http://localhost:55555
export NAMESPACE=foo
export K_METRICS_CONFIG='{"config-observability":"{\"metrics.backend\":\"prometheus\"}"}'
export K_LOGGING_CONFIG='{"zap-logger-config": "{\"level\":\"info\",\"development\":\"true\",\"outputPaths\":[\"stdout\"],\"errorOutputPaths\":[\"stderr\"],\"encoding\":\"console\",\"encoderConfig\":{\"timeKey\":\"ts\",\"levelKey\":\"level\",\"nameKey\":\"logger\",\"callerKey\":\"caller\",\"messageKey\":\"msg\",\"stack traceKey\":\"stacktrace\",\"lineEnding\":\"\",\"levelEncoder\":\"\",\"timeEncoder\":\"iso8601\",\"durationEncoder\":\"\",\"callerEncoder\":\"\"}}"}'
export EVENT_SOURCE="varkes"
```

2. Run the adapter:

```bash
go run cmd/http-adapter/main.go
```

As a result, the adapter will send events to the `SINK_URI`. When running the adapter locally, you can simply use `netcat` as a sink:

```bash
printf "HTTP/1.1 200 OK\r\n\r\n" | nc -vl 55555
```

### Send events to the adapter

You can send events using the structured or binary content mode.


To use the structured content mode, run:

```bash
curl -v -d '{
    "specversion" : "1.0",
    "type": "foo",
    "source": "will be replaced",
    "type": "foo",
    "eventtypeversion": "0.3",
    "id" : "A234-1234-1234",
    "datacontenttype" : "text/xml",
    "data" : "<much wow=\"xml\"/>"
}' -H "Content-Type: application/cloudevents+json" -X POST http://localhost:8080
```

To use the binary content mode, run:

```bash
curl -v \
    -H "ce-specversion: 1.0" \
    -H "ce-type: foo" \
    -H "ce-source: will be replaced" \
    -H "ce-type: foo" \
    -H "ce-eventtypeversion: 0.3" \
    -H "ce-id: A234-1234-1234" \
    -H "content-type: text/xml" \
    -d '<much wow=\"xml\"/>' \
    http://localhost:8080
```

## Development

### Testing

Run unit and integration tests:

```bash
make test-local
# or: go test ./adapter/http/
```
