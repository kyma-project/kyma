---
title: Use the OpenTelemetry runtime client
---

This tutorial shows how to use the build-in OpenTelemetry client to send trace data to the Jaeger service.

Function runtimes are configured to handle traced requests, propagate right headers to the function contexts, and automatically decorate outgoing requests by right headers. The [event](../../05-technical-reference/svls-08-function-specification.md#event-object) object contains necessary tools with allow to pass additional information about trace by creating custom [spans](https://opentelemetry.io/docs/reference/specification/trace/api/#span) and adding data to them like events. This tutorial shows how to do this.

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
         span = event.tracer.startSpan('foo', {});
         span.addEvent('bar');
         span.end();
         
         return "Hello OpenTelemetry"
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
      with event.tracer.start_as_current_span("fir-span"):
         return "hello world" 
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
