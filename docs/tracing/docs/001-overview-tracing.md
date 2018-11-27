---
title: Overview
type: Overview
---

The microservice architecture differs from the traditional monoliths in many aspects. From the request observability perspective, there are asynchronous boundaries among various different microservices that compose a request flow. Moreover, these microservices can have heterogeneous semantics when it comes to monitoring. A tracing solution that provides a holistic view of the request flow helps you to understand the system and take informed decisions regarding troubleshooting and performance optimization.

Tracing in Kyma uses [Jaeger](https://www.jaegertracing.io/docs/) as a backend which serves as the query mechanism for displaying information about traces.
