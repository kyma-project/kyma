---
title: Overview
type: Overview
---

The micro-services based architecture differs from the traditional monoliths in many aspects. From the request observability perspective, there are asynchronous boundaries among various different micro-services that compose a request flow. Moreover, these micro-services can have heterogeneous semantics when it comes to monitoring and observability. It is required to have a tracing solution that can provide a holistic view of the request flow and help the developer understand the system better to take informed decisions regarding troubleshooting and performance optimization.

Tracing in Kyma uses [Jaeger](https://www.jaegertracing.io/docs/) as a backend which serves as the query mechanism for displaying information about traces. Jaeger is used for monitoring and troubleshooting microservice-based distributed systems, including:

- Distributed context propagation
- Distributed transaction monitoring
- Root cause analysis
- Service dependency analysis
- Performance and latency optimization

Jaeger provides compatibility with the Zipkin protocol. The compatibility makes it possible to use Zipkin protocol and clients in Istio, Envoy, and Kyma services.

## Access Jaeger

To access the Jaeger UI, follow these steps:

1. Run the following command to configure port-forwarding:

```
kubectl port-forward -n kyma-system $(kubectl get pod -n kyma-system -l app=jaeger -o jsonpath='{.items[0].metadata.name}') 16686:16686
```

2. Access the Jaeger UI at `http://localhost:16686`.

## Propagate HTTP headers

The envoy proxy controls the inbound and outbound traffic in the application and automatically sends the trace information to the Zipkin. However, to track the flow of the REST API calls or the service injections in Kyma, it requires the minimal application cooperation from the micro-services code. For this purpose, you need to configure the application to propagate the tracing context in HTTP headers when making outbound calls. See the [Istio documentation](https://istio.io/docs/tasks/telemetry/distributed-tracing.html#understanding-what-happened) for details on which headers are required to ensure the correct tracing in Kyma.
