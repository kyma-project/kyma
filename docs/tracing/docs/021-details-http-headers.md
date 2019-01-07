---
title: Propagate HTTP headers
type: Details
---

The Envoy proxy controls the inbound and outbound traffic in the application and automatically sends the trace information to Zipkin. To track the flow of the REST API calls or the service injections in Kyma, it requires the application to cooperate with the microservices code. To enable such cooperation, configure the application to propagate the tracing context in HTTP headers when making outbound calls. See the [Istio documentation](https://istio.io/docs/tasks/telemetry/distributed-tracing.html#understanding-what-happened) for details on headers required to ensure the correct tracing in Kyma.
