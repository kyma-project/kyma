---
title: What is Observability in Kyma?
---

Kyma comes with tools that give you the most accurate and up-to-date monitoring, logging and tracing data.

- [Prometheus](https://prometheus.io/) is the open source monitoring and alerting toolkit that provides the telemetry data. This data is consumed by different add-ons, including [Grafana](https://grafana.com/) for analytics and monitoring, and [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) for handling alerts.
- For logging, Kyma uses [Loki](https://github.com/grafana/loki), a Prometheus-like log management system.
- With the [Jaeger](https://github.com/jaegertracing) distributed tracing system, you can analyze the path of a request chain going through your distributed applications. This information helps you, for example, to troubleshoot your applications, or optimize the latency and performance of your solution.
