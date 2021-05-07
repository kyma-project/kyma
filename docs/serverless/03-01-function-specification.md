---
title: Function's specification
type: Details
---

Serverless in Kyma allows you to create Functions in both Node.js (v12 & v14) and Python (v3.8). Although the Function's interface is unified, its specification differs depending on the runtime used to run the Function.

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
    "data": b"",
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
| **extensions** | JSON object that can contain event payload, a Function's incoming request, or an outgoing response |

### Context object

The `context` object contains information about the Function invocation, such as runtime details, execution timeout, or memory limits.

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
| **runtime** | Environment used to run the Function. You can use `nodejs12`, `nodejs14`, or `python3.8`. |
| **memory-limit** | Maximum amount of memory assigned to run a Function |  

## HTTP requests

You can use the **event.extensions.request** object to access properties and methods of a given request that vary depending on the runtime. For more information, read the API documentation for [Node.js](https://nodejs.org/docs/latest-v12.x/api/http.html#http_class_http_clientrequest) and [Python](https://bottlepy.org/docs/dev/api.html#the-request-object).

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
