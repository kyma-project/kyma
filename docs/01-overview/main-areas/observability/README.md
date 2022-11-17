---
title: What is Observability in Kyma?
---

Fundamentally, "Observability" is a measure of how well the internal states of single components can be reflected by the application's external outputs. The insights that an application exposes are displayed in the form of metrics, traces, and logs - collectively, that's called "telemetry" or "signals". These can be exposed by employing modern instrumentation.

Out of the box, Kyma provides tools to collect and expose telemetry data. Of course, you'll want to view and analyze the data you're collecting. This is where observability tools come in.

## Data collection

Kyma collects telemetry data with the following in-cluster components:

- [Prometheus](https://prometheus.io/docs/introduction) collects metrics from Pods. Metrics are the time-stamped data that provide information on the running jobs, workload, CPU consumption, memory usage, and more. All metrics relevant for observing the in-cluster Istio Service Mesh are collected separately.

- [Fluent Bit](https://fluentbit.io/) collects logs.

- Traces are sent to [Jaeger](https://www.jaegertracing.io/docs).

The collected telemetry data are exposed so that you can view and analyze them with observability tools.

> **NOTE:** Kyma's [telemetry component](./obsv-04-telemetry-in-kyma.md) supports providing your own output configuration for application logs. With this, you can connect your own observability systems outside the Kyma cluster with the Kyma backend.

## Data analysis

You can use the following in-cluster components to observe your applications' telemetry data:

- [Prometheus](https://prometheus.io/docs/introduction), a lightweight backend for metrics.
- [Jaeger](https://www.jaegertracing.io/docs/), a tracing backend serving as the query mechanism to display information about traces.

- [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) to receive and manage alerts coming from Prometheus. Alertmanager can then forward the notifications about fired alerts to specific channels, such as Slack or an on-call paging system of your choice.
- [Grafana](https://grafana.com/docs/guides/getting_started/) to provide a dashboard and a query editor to visualize metrics and logs collected from Prometheus and Loki.
