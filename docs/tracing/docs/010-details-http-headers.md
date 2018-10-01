---
title: Propagate HTTP headers
type: Details
---

The envoy proxy controls the inbound and outbound traffic in the application and automatically sends the trace information to the Zipkin. However, to track the flow of the REST API calls or the service injections in Kyma, it requires the minimal application cooperation from the micro-services code. For this purpose, you need to configure the application to propagate the tracing context in HTTP headers when making outbound calls. See the [Istio documentation](https://istio.io/docs/tasks/telemetry/distributed-tracing.html#understanding-what-happened) for details on which headers are required to ensure the correct tracing in Kyma.
