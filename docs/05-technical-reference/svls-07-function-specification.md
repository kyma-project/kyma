---
title: Function's specification
---

Serverless in Kyma allows you to create Functions in both Node.js and Python. Although the Function's interface is unified, its specification differs depending on the runtime used to run the Function.

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
| **tracer** | Fully configured OpenTelemetry [tracer](https://opentelemetry.io/docs/reference/specification/trace/api/#tracer) object that allows you to communicate with the user-defined trace backend service to share tracing data. For more information on how to use the tracer object see [Customize Function traces](../03-tutorials/00-serverless/svls-12-customize-function-traces.md) |
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
| **publishCloudEvent** | event | **Deprecated: use `emitCloudEvent` instead.** Publishes a CloudEvent on the publisher service based on the given CloudEvent object |
| **buildResponseCloudEvent** | id, type, data | **Deprecated: use `emitCloudEvent` instead.** Builds a CloudEvent object based on the `request` CloudEvent object and the given arguments |
| **emitCloudEvent** | type, source, data, optionalCloudEventAttribute | Builds a CloudEvent based on the arguments and emits it on the eventing publisher service. You can pass any additional [cloudevent attributes](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/formats/json-format.md#2-attributes) as properties of the last optional argument `optionalCloudEventAttribute` |

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
  "runtime": "nodejs18",
  "memory-limit": 200Mi }
```

See the detailed descriptions of these fields:

| Field | Description                                                                               |
|-------|-------------------------------------------------------------------------------------------|
| **function-name** | Name of the invoked Function                                                              |
| **timeout** | Time, in seconds, after which the system cancels the request to invoke the Function       |
| **runtime** | Environment used to run the Function. You can use `nodejs16` - deprecated, `nodejs18`, or `python39`. |
| **memory-limit** | Maximum amount of memory assigned to run a Function                                       |

## HTTP requests

You can use the **event.extensions.request** object to access properties and methods of a given request that vary depending on the runtime. For more information, read the API documentation for [Node.js Express](http://expressjs.com/en/api.html#req) and [Python](https://bottlepy.org/docs/dev/api.html#the-request-object).

## Custom HTTP responses

<div tabs name="custom-http-response" group="function-specification">
<details>
<summary label="Node.js">
Node.js
</summary>

By default, a failing Function simply throws an error to tell the Event Service to reinject the event at a later point. Such an HTTP-based Function returns the HTTP status code `500`. On the contrary, if you manage to invoke a Function successfully, the system returns the default HTTP status code `200`.

Apart from these two default codes, you can define custom responses in all Node.js runtimes using the **event.extensions.response** object.

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

</details>
<details>
<summary label="Python">
Python
</summary>

By default, a failing Function simply throws an error to tell the Event Service to reinject the event at a later point. Such an HTTP-based Function returns the HTTP status code `500`. In case you manage to invoke a Function successfully, the system returns the default HTTP status code `200`.

Apart from these two default codes, you can define custom responses in all Python runtimes using the **HTTPResponse** object available in Bottle.

This object will need to be instantiated and can be customized. For more information, read [Bottle API documentation](https://bottlepy.org/docs/dev/api.html#the-response-object).

The following example shows how to set such a custom response in Python for the HTTP status code `400`:

```python
from bottle import HTTPResponse

SUPPORTED_CONTENT_TYPES = ['application/json']

def main(event, context):
    request = event['extensions']['request']

    response_content_type = 'application/json'
    headers = {
        'Content-Type': response_content_type
    }

    status = 202
    response_payload = {'success': 'Message accepted.'}

    if request.headers.get('Content-Type') not in SUPPORTED_CONTENT_TYPES:
        status = 400
        response_payload = json.dumps({'error': 'Invalid Content-Type.'})

    return HTTPResponse(body=response_payload, status=status, headers=headers)
```

</details>
</div>

## /metrics endpoint  

You can use the `/metrics` endpoint to return the Function metrics. All the information is gathered using Prometheus and can be displayed using the Grafana dashboard (see [Kyma observability](../02-get-started/05-observability.md) for more information on how to use Grafana dashboard in Kyma). As this endpoint is provided by Kubeless, it cannot be customized.  
For more information, see [Kubeless monitoring](https://github.com/vmware-archive/kubeless/blob/master/docs/monitoring.md) and [Kubeless runtime variants](https://github.com/vmware-archive/kubeless/blob/master/docs/runtimes.md) pages.

> **NOTE:** Prometheus and Grafana are [deprecated](https://kyma-project.io/blog/2022/12/9/monitoring-deprecation) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

## Override runtime image

You can use a custom runtime image to override the existing one. Your image must meet all the following requirements:

- Expose the workload endpoint on the right port
- Provide liveness and readiness check endpoints at `/healthz`
- Fetch sources from the path under the `KUBELESS_INSTALL_VOLUME` environment
- Security support. Kyma runtimes are secure by default. You only need to protect your images.

> Note: For better understanding you can look at [main dockerfiles](https://github.com/kyma-project/kyma/blob/main/resources/serverless/templates/runtimes.yaml) which are responsible for building the final image based on the `base_image` argument which you as an user can override and what we are doing in [this tutorial](../03-tutorials/00-serverless/svls-13-override-runtime-image).

Every Function's Pods container have the same system environments which helps you configure the Functions server. For more information, read the [Environment variables](../05-technical-reference/00-configuration-parameters/svls-02-environment-variables.md) page.
