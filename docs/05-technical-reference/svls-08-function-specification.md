---
title: Function's specification
---

Serverless in Kyma allows you to create Functions in both Node.js (v12 & v14) and Python (v3.9). Although the Function's interface is unified, its specification differs depending on the runtime used to run the Function.

## Signature

Function's code is represented by a handler that is a method that processes events. When your Function is invoked, it runs this handler method to process a given request and return a response.

All Functions have a predefined signature with elements common for all available runtimes:
- Functions' code must be introduced by the `main` handler name.
- Functions must accept two arguments that are passed to the Function handler:
    - `event`
    - `context`

See these signatures for each runtime:

<div tabs name="signature" group="function-specification">
  <details>
  <summary label="Node.js">
  Node.js
  </summary>

```bash
module.exports = {
  main: function (event, context) {
    return
  }
}
```

</details>
<details>
<summary label="Python">
Python
</summary>

```bash
def main(event, context):
    return
```

</details>
</div>

### Event object

The `event` object contains information about the event the Function refers to. For example, an API request event holds the HTTP request object.

Functions in Kyma accept [CloudEvents](https://cloudevents.io/) (**ce**) with the following specification:

<div tabs name="signature" group="function-specification">
  <details>
  <summary label="Node.js">
  Node.js
  </summary>

```json
...
{
    "ce-type": "com.github.pull_request.opened",
    "ce-source": "/cloudevents/spec/pull/123",
    "ce-eventtypeversion": "v1",
    "ce-specversion": "1.0",
    "ce-id": "abc123",
    "ce-time": "2020-12-20T13:37:33.647Z",
    "data": {BUFFER},
    "tracer": {OPENTELEMETRY_TRACER},
    "extensions": {
        "request": {INCOMING_MESSAGE},
        "response": {SERVER_RESPONSE},
    }
}
```
</details>
<details>
<summary label="Python">
Python
</summary>

```json
{
    "ce-type": "com.github.pull_request.opened",
    "ce-source": "/cloudevents/spec/pull/123",
    "ce-eventtypeversion": "v1",
    "ce-specversion": "1.0",
    "ce-id": "abc123",
    "ce-time": "2020-12-20T13:37:33.647Z",
    "data": "",
    "tracer": {OPENTELEMETRY_TRACER},
    "extensions": {
        "request": {PICKLABLE_BOTTLE_REQUEST},
    }
}
```

</details>
</div>

See the detailed descriptions of these fields:

| Field | Description |
|-------|-------------|
| **ce-type** | Type of the CloudEvent data related to the originating occurrence |
| **ce-source** | Unique context in which an event happened and can relate to an organization or a process |
| **ce-eventtypeversion** | Version of the CloudEvent type |
| **ce-specversion** | Version of the CloudEvent specification used for this event |
| **ce-id** | Unique identifier of the event |
| **ce-time** | Time at which the event was sent |
| **data** | Either JSON or a string, depending on the request type. Read more about [Buffer](https://nodejs.org/api/buffer.html) in Node.js and [bytes literals](https://docs.python.org/3/reference/lexical_analysis.html#string-and-bytes-literals) in Python. |
| **tracer** | Fully configured OpenTelemetry [tracer](https://opentelemetry.io/docs/reference/specification/trace/api/#tracer) object that allows you to communicate with the Jaeger service to share tracing data. For more information on how to use the tracer object see [Use the OpenTelemetry standard](../03-tutorials/00-serverless/svls-12-use-opentelemetry-client.md) |
| **extensions** | JSON object that can contain event payload, a Function's incoming request, or an outgoing response |

### Event object SDK

The `event` object is extended by methods making some operations easier. You can use every method by providing `event.{FUNCTION_NAME(ARGUMENTS...)}`.

<div tabs name="signature" group="function-specification">
<details>
<summary label="Node.js">
Node.js
</summary>

| Method name | Arguments | Description |
|---------------|-----------|-------------|
| **setResponseHeader** | key, value | Sets a header to the `response` object based on the given key and value |
| **setResponseContentType** | type | Sets the `ContentType` header to the `response` object based on the given type |
| **setResponseStatus** | status | Sets the `response` status based on the given status |
| **publishCloudEvent** | event | Publishes a CloudEvent on the publisher service based on the given CloudEvent object |
| **buildResponseCloudEvent** | id, type, data | Builds a CloudEvent object based on the `request` CloudEvent object and the given arguments |

</details>
<details>
<summary label="Python">
Python
</summary>

| Method name | Arguments | Description |
|----------|-----------|-------------|
| **publishCloudEvent** | event | Publishes a CloudEvent on the publisher service based on the given CloudEvent object |
| **buildResponseCloudEvent** | id, type, data | Builds a CloudEvent object based on the `request` CloudEvent object and the given arguments |

</details>
</div>

### Context object

The `context` object contains information about the Function's invocation, such as runtime details, execution timeout, or memory limits.

See sample context details:

```json
...
{ "function-name": "main",
  "timeout": 180,
  "runtime": "nodejs14",
  "memory-limit": 200Mi }
```

See the detailed descriptions of these fields:

| Field | Description |
|-------|-------------|
| **function-name** | Name of the invoked Function |
| **timeout** | Time, in seconds, after which the system cancels the request to invoke the Function |
| **runtime** | Environment used to run the Function. You can use `nodejs14` or `python39`. |
| **memory-limit** | Maximum amount of memory assigned to run a Function |

## HTTP requests

You can use the **event.extensions.request** object to access properties and methods of a given request that vary depending on the runtime. For more information, read the API documentation for [Node.js Express](http://expressjs.com/en/api.html#req) and [Python](https://bottlepy.org/docs/dev/api.html#the-request-object).

## Custom HTTP responses in Node.js

By default, a failing Function simply throws an error to tell the event service to reinject the event at a later point. Such an HTTP-based Function returns the HTTP status code `500`. On the contrary, if you manage to invoke a Function successfully, the system returns the default HTTP status code `200`.

Apart from these two default codes, you can define custom responses in both Node.js 12 and Node.js 14 environments using the **event.extensions.response** object.

This object is created by the Express framework and can be customized. For more information, read [Node.js API documentation](https://nodejs.org/docs/latest-v12.x/api/http.html#http_class_http_serverresponse).

This example shows how to set such a custom response in Node.js for the HTTP status code `400`:

```js
module.exports = {    
    main: function (event, context) {
        if (event.extensions.request.query.id === undefined) {
            res = event.extensions.response;
            res.status(400)
            return
        }
        return "42"
    }
}
```

## /metrics endpoint  

You can use the `/metrics` endpoint to return the Function metrics. All the information is gathered using Prometheus and can be displayed using the Grafana dashboard (see [Kyma observability](https://kyma-project.io/docs/kyma/latest/02-get-started/05-observability/) for more information on how to use Grafana dashboard in Kyma). As this endpoint is provided by Kubeless, it cannot be customized.  
For more information, see [Kubeless monitoring](https://github.com/vmware-archive/kubeless/blob/master/docs/monitoring.md) and [Kubeless runtime variants](https://github.com/vmware-archive/kubeless/blob/master/docs/runtimes.md) pages.
