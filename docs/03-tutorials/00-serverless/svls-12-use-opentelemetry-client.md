---
title: Use the OpenTelemetry standard
---

This tutorial shows how to use the build-in OpenTelemetry tracer object to send custom trace data to the Jaeger service.

Kyma Functions are instrumented to handle trace headers. This means that every time you call your Function, the executed logic is traceable using a dedicated span visible in the Jaeger service (i.e. start time and duration).
Additionally, you can extend the default trace context and create your own custom spans whenever you feel it is helpful (i.e. when calling a remote service in your distributed application) or add additional information to the tracing context by introducing events and tags. This tutorial shows you how to do it using tracer client that is available as part of the [event](../../05-technical-reference/svls-08-function-specification.md#event-object) object.

## Prerequisites

Before you start, make sure you have these tools installed:

- Kyma installed on a cluster

## Steps

Follows these steps:

1. [Create an inline Function](./svls-01-create-inline-function.md) with the following body:

   <div tabs name="code" group="functions-code">
   <details>
   <summary label="node.js">
   Node.js
   </summary>

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

   </details>
   <details>
   <summary label="python">
   Python
   </summary>

      ```python
      def main(event, context):
         span = event.tracer.start_span("foo")
         span.add_event("bar")
         span.end()

         return "hello OpenTelemetry"
      ```

   </details>
   </div>

2. [Expose your Function](./svls-03-expose-function.md).
3. [Expose Jaeger securely](../../04-operation-guides/security/sec-06-access-expose-kiali-grafana.md) and open the following Jaeger's address in your browser:

   ```text
   http://localhost:16686
   ```

   > **NOTE:** By default, only 1% of the requests are sent to Jaeger for the trace recording. To increase this number see the [Jaeger doesn't show the traces you want to see](../../04-operation-guides/troubleshooting/observability/obsv-02-troubleshoot-jaeger-shows-few-traces.md) page.

4. Choose your Deployment's name from the **Service** list and click **Find Traces**
