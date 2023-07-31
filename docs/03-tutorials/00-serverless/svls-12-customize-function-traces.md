---
title: Customize Function traces
---

This tutorial shows how to use the build-in OpenTelemetry tracer object to send custom trace data to the trace backend.

Kyma Functions are instrumented to handle trace headers. This means that every time you call your Function, the executed logic is traceable using a dedicated span visible in the trace backend (i.e. start time and duration).
Additionally, you can extend the default trace context and create your own custom spans whenever you feel it is helpful (i.e. when calling a remote service in your distributed application) or add additional information to the tracing context by introducing events and tags. This tutorial shows you how to do it using tracer client that is available as part of the [event](../../05-technical-reference/svls-07-function-specification.md#event-object) object.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Kyma installed](../../04-operation-guides/operations/02-install-kyma.md) on a cluster
- Trace backend configured to collect traces from the cluster. You can bring your own trace backend or deploy [Jaeger](https://github.com/kyma-project/examples/tree/main/jaeger).

## Steps

The following code samples illustrate how to enrich the default trace with custom spans, events, and tags:

1. [Create an inline Function](./svls-01-create-inline-function.md) with the following body:

   <div tabs name="code" group="functions-code">
   <details>
   <summary label="node.js">
   Node.js
   </summary>

      ```javascript

      const { SpanStatusCode } = require("@opentelemetry/api/build/src/trace/status");
      const axios = require("axios")
      module.exports = {
         main: async function (event, context) {

            const data = {
               name: "John",
               surname: "Doe",
               type: "Employee",
               id: "1234-5678"
            }

            const span = event.tracer.startSpan('call-to-acme-service');
            return await callAcme(data)
               .then(resp => {
                  if(resp.status!==200){
                    throw new Error("Unexpected response from acme service");
                  }
                  span.addEvent("Data sent");
                  span.setAttribute("data-type", data.type);
                  span.setAttribute("data-id", data.id);
                  span.setStatus({code: SpanStatusCode.OK});
                  return "Data sent";
               }).catch(err=> {
                  console.error(err)
                  span.setStatus({
                    code: SpanStatusCode.ERROR,
                    message: err.message,
                  });
                  return err.message;
               }).finally(()=>{
                  span.end();
               });
         }
      }

      let callAcme = (data)=>{
         return axios.post('https://acme.com/api/people', data)
      }
      ```

   </details>
   <details>
   <summary label="python">
   Python
   </summary>

      [OpenTelemetry SDK](https://opentelemetry.io/docs/instrumentation/python/manual/#traces) allows you to customize trace spans and events.
      Additionally, if you are using the `requests` library then all the HTTP communication can be auto-instrumented:

      ```python

      import requests
      import time
      from opentelemetry.instrumentation.requests import RequestsInstrumentor

      def main(event, context):
         # Create a new span to track some work
         with event.tracer.start_as_current_span("parent"):
            time.sleep(1)

            # Create a nested span to track nested work
            with event.tracer.start_as_current_span("child"):
               time.sleep(2)
               # the nested span is closed when it's out of scope

         # Now the parent span is the current span again
         time.sleep(1)

         # This span is also closed when it goes out of scope

         RequestsInstrumentor().instrument()

         # This request will be auto-intrumented
         r = requests.get('https://swapi.dev/api/people/2')
         return r.json()
      ```

   </details>
   </div>

2. [Expose your Function](./svls-03-expose-function.md).

3. Find the traces for the Function in the trace backend.
