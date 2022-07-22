---
title: Tracing
---

With the [Jaeger](https://github.com/jaegertracing) distributed tracing system, you can analyze the path of a request chain going through your distributed applications. This information helps you to, for example, troubleshoot your applications, or optimize the latency and performance of your solution.

## Limitations

In the [production profile](../../../04-operation-guides/operations/02-install-kyma.md#choose-resource-consumption), Jaeger has no persistence enabled and keeps up to 10.000 traces stored in-memory. The oldest records are removed first. The evaluation profile has lower limits.

## Benefits of distributed tracing

Observability tools should clearly show the big picture, no matter if you're monitoring just a few or many components. In a cloud-native microservice architecture, a user request often flows through dozens of different microservices. Logging and monitoring tools help to track the request's path. However, they treat each component or microservice in isolation. This individual treatment results in operational issues.

Distributed tracing charts out the transactions in cloud-native systems, helping you to understand the application behavior and relations between the frontend actions and backend implementation.

The diagram shows how distributed tracing helps to track the request path:

![Distributed tracing](./assets/distributed-tracing.svg)
