---
title: What is Observability in Kyma?
---

Kyma comes with tools that give you the most accurate and up-to-date monitoring, logging and tracing data.

- [Prometheus](https://prometheus.io/) is the open source monitoring and alerting toolkit that provides the telemetry data. This data is consumed by different addons, including [Grafana](https://grafana.com/) for analytics and monitoring, and [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) for handling alerts.
- For logging, Kyma uses [Loki](https://github.com/grafana/loki), a Prometheus-like log management system.
- With the [Jaeger](https://github.com/jaegertracing) distributed tracing system, you can analyze the path of a request chain going through your distributed applications. This information helps you to, for example, troubleshoot your applications, or optimize the latency and performance of your solution.

## Benefits of distributed tracing

Observability tools should clearly show the big picture, no matter if you're monitoring just a couple or multiple components. In a cloud-native microservice architecture, a user request often flows through dozens of different microservices. Tools such as logging or monitoring help to track the way, however, they treat each component or microservice in isolation. This individual treatment results in operational issues.

Distributed tracing charts out the transactions in cloud-native systems, helping you to understand the application behavior and relations between the frontend actions and backend implementation.

The diagram shows how distributed tracing helps to track the request path.

![Distributed tracing](./assets/distributed-tracing.svg)
