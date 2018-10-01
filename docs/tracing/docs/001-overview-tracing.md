---
title: Overview
---

The micro-services based architecture differs from the traditional monoliths in many aspects. From the request observability perspective, there are asynchronous boundaries among various different micro-services that compose a request flow. Moreover, these micro-services can have heterogeneous semantics when it comes to monitoring and observability. It is required to have a tracing solution that can provide a holistic view of the request flow and help the developer understand the system better to take informed decisions regarding troubleshooting and performance optimization.

Tracing in Kyma uses [Jaeger](https://www.jaegertracing.io/docs/) as a backend which serves as the query mechanism for displaying information about traces. Jaeger is used for monitoring and troubleshooting microservice-based distributed systems, including:

- Distributed context propagation
- Distributed transaction monitoring
- Root cause analysis
- Service dependency analysis
- Performance and latency optimization

Jaeger provides compatibility with the Zipkin protocol. The compatibility makes it possible to use Zipkin protocol and clients in Istio, Envoy, and Kyma services.

You can access the Jaeger UI either locally at `https://jaeger.kyma.local` or on a cluster at `https://jaeger.{domain-of-kyma-cluster}`. 
