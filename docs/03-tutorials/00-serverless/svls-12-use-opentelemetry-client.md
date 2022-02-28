---
title: Use the OpenTelemetry runtime client
---

This tutorial shows how to use the build-in OpenTelemetry client to send custom trace data to the Jaeger service.

Kyma Functions are instrumented to handle trace headers. That means every time your function is called, the executed logic is traceable via dedicated span visible in the Jeager service  ( i.e start time and duration ). 
Additionally, you can extend the default trace context and create our own custom spans whenever you feel it's helpful (i.e when calling a remote service in your distributed application ) or add additional information to the tracing context by introducing events and tags. This tutorial shows how to do it via tracer client that is available as part of   [event](../../05-technical-reference/svls-08-function-specification.md#event-object) object.

## Steps

Follows these steps:

<div tabs name="steps" group="opentelemetry-client">
  <details>
  <summary label="node.js">
  Node.js
  </summary>

1. [Create inline function](./svls-01-create-inline-function.md) with following body:

   ```javascript
   module.exports = {
      main: function (event, context) {
         span = event.tracer.startSpan('foo');
         span.addEvent('bar');
         span.end();
         
         return "hello OpenTelemetry"
      }
   }
   ```

2. [Expose function](./svls-03-expose-function.md) and access the Function's external address.
3. [Expose Jaeger securely](../../04-operation-guides/security/sec-06-access-expose-kiali-grafana.md).
4. Open the following Jaeger's addres in the browser:

   ```text
   http://localhost:16686
   ```

5. Find and select the deployment's name in the `Service` list and click `Find Traces`.

</details>
<details>
<summary label="python">
Python
</summary>

1. [Create inline function](./svls-01-create-inline-function.md) with following body:

   ```python
   def main(event, context):
      span = event.tracer.start_span("foo")
      span.add_event("bar")
      span.end()

      return "hello OpenTelemetry"
   ```

2. [Expose function](./svls-03-expose-function.md) and access the Function's external address.
3. [Expose Jaeger securely](../../04-operation-guides/security/sec-06-access-expose-kiali-grafana.md).
4. Open the following Jaeger's addres in the browser:

   ```text
   http://localhost:16686
   ```

5. Find and select the deployment's name in the `Service` list and click `Find Traces`.

</details>
</div>
