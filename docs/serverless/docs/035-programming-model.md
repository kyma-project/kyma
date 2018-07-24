---
title: The Node.js Programming Model
type: Model
---

## Overview

Kyma supports Node.js 6 and 8. The function interface is the same for both versions. It is still best practice to start with Node.js 8, as it supports Promises out of the box. The result is less complicated code.

Please set the runtime version (Node.js 6 or 8) while creating a function.

In the next sections, we will describe how the system creates Node.js functions.

### The Handler

The system uses ```module.exports``` to export Node.js handlers. A handler represents the function code executed during invocation. You have to define the handler using the command line. The Console UI only supports ```main``` as a handler name.

```JavaScript
module.exports = { main: function (event, context) {
    return
} }
```

Kyma  supports two execution types: **Request / Response (HTTP)** and **Events**. In both types, a ```return``` identifies a successful execution of the function. For event types, the event is reinjected as long as the execution is not successful. Functions of the Request Response type can return data to the requesting entity. The following three options are available:

| Return                      | Content Type     | HTTP Status | Response      |
| --------------------------- | ---------------- | ----------- | ------------- |
| ```return```                | none             | 200 (OK)    | -             |
| ```return "Hello World!"``` | none             | 200 (OK)    | Hello World!  |
| ```return {foo: "BAR"}```   | application/json | 200 (OK)    | {"foo":"BAR"} |

A failing function simply throws an error to tell the event service to reinject the event at a later point. An HTTP-based function returns an HTTP 500 status.

### The Event Object and Context Object

The function retrieves two parameters: Event and Context.

```yaml
event:
  data:                                         # Event data
    foo: "bar"                                  # The data is parsed as JSON when required
  extensions:                                   # Optional parameters
    request: ...                                # Reference to the request received 
    response: ...                               # Reference to the response to send 
                                                # (specific properties will depend on the function language)
context:
    function-name: "pubsub-nodejs"
    timeout: "180"
    runtime: "nodejs6"
    memory-limit: "128M"
```

The Event contains the event payload as well as some request specific metadata. The request and response attributes are primarily responsible for providing control over http behavior.

### Advanced Response Handling

To enable more advanced implementations, the system forwards Node.js Request and Response objects to the function. Access the objects using ```event.extensions.<request|response>```.

In the example, a custom HTTP response is set.

```JavaScript
module.exports = { main: function (event, context) {
    console.log(event.extensions.request.originalUrl)
    event.extensions.response.status(404).send("Arg....")
} }
```

The example code logs the original request url. The response is an HTTP 404. The body is ```Arg....```.

### Logging

Logging is based on standard Node.js functionality. ```console.log("Hello")``` sends "Hello" to the logs. As there is no graphical log tool available, use the command ```kubectl``` to display the logs.

```sh
$ kubectl logs -n <environment> -l function=<function> -c <function>
```