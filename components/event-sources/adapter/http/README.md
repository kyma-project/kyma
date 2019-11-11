# HTTP Adapter Source

The http adapter is a http server which receives [CloudEvents](https://github.com/cloudevents/spec) in version 1.0 and proxies them to a preconfigured sink.
The adapter is written with the [cloudevents sdk-go](https://github.com/cloudevents/sdk-go) and is fully compatible with CloudEvents specification.

## Specification

The http adapter is CloudEvents 1.0 compatible. It accepts binary as well as structured content mode. Only batched mode is not implemented.
Read more about the content types and the http protocol binding [here](https://github.com/cloudevents/spec/blob/master/http-protocol-binding.md#13-content-modes).

## Run

The adapter can be run locally.

It requires a few variables to be set:

```bash
export SINK_URI=http://localhost:55555
export NAMESPACE=foo
export K_METRICS_CONFIG='{"config-observability":"{\"metrics.backend\":\"prometheus\"}"}'
export K_LOGGING_CONFIG='{"zap-logger-config": "{\"level\":\"info\",\"development\":\"true\",\"outputPaths\":[\"stdout\"],\"errorOutputPaths\":[\"stderr\"],\"encoding\":\"console\",\"encoderConfig\":{\"timeKey\":\"ts\",\"levelKey\":\"level\",\"nameKey\":\"logger\",\"callerKey\":\"caller\",\"messageKey\":\"msg\",\"stack traceKey\":\"stacktrace\",\"lineEnding\":\"\",\"levelEncoder\":\"\",\"timeEncoder\":\"iso8601\",\"durationEncoder\":\"\",\"callerEncoder\":\"\"}}"}'
export EVENT_SOURCE="varkes"
```

Run it:
```bash
go run cmd/http-adapter/main.go
```

The adapter will send events to the `SINK_URI`. When running the adapter locally, you can simple use `netcat` as a sink:
```bash
printf "HTTP/1.1 200 OK\r\n\r\n" | nc -vl 55555
```

Send an event to the adapter:

```bash
# structured mode
$ curl -v -d '{
	"specversion" : "1.0",
	"type": "foo",
	"source": "will be replaced",
	"type": "foo",
	"eventtypeversion": "0.3",
	"id" : "A234-1234-1234",
    "datacontenttype" : "text/xml",
	"data" : "<much wow=\"xml\"/>"
}' -H "Content-Type: application/cloudevents+json" -X POST http://localhost:8080

# binary mode
$ curl -v \
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

## Testing

Unit and integration tests can be run using the following command:

```bash
make test
# or: go test ./adapter/http/
```