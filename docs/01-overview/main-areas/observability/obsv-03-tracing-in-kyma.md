---
title: Tracing in Kyma
---

With the [Jaeger](https://github.com/jaegertracing) distributed tracing system, you can analyze the path of a request chain going through your distributed applications. This information helps you to, for example, troubleshoot your applications, or optimize the latency and performance of your solution.

## Limitations

By default, in the production profile, Jaeger has no persistence enabled and keeps all data in-memory with a retention of 10.000 traces.

The evaluation profile has lower limits. For more information about profiles, see [Install Kyma: Choose resource consumption](../../../04-operation-guides/operations/02-install-kyma.md#choose-resource-consumption).


## Benefits of distributed tracing

Observability tools should clearly show the big picture, no matter if you're monitoring just a few or many components. In a cloud-native microservice architecture, a user request often flows through dozens of different microservices. Tools such as logging or monitoring help to track the request's path. However, they treat each component or microservice in isolation. This individual treatment results in operational issues.

Distributed tracing charts out the transactions in cloud-native systems, helping you to understand the application behavior and relations between the frontend actions and backend implementation.

The diagram shows how distributed tracing helps to track the request path.

![Distributed tracing](./assets/distributed-tracing.svg)
