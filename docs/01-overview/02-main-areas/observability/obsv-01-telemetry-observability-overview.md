---
title: Telemetry and Observability
---

Out of the box, Kyma provides tools to collect and expose **telemetry** data, such as metrics, traces, and log data. Of course, you'll want to view and analyze the data you're collecting. This is where **observability** tools come in.

## Collecting data

Kyma collects telemetry data with several in-cluster components:

[Prometheus](https://prometheus.io/docs/introduction) collects metrics from Pods. Metrics are the time-stamped data that provide information on the running jobs, workload, CPU consumption, memory usage, and more.

> **NOTE:** All metrics relevant for observing the in-cluster Istio Service Mesh are collected separately. You can find more information about it in the [Istio monitoring documentation](../../../01-overview/02-main-areas/service-mesh/con-monitoring-istio.md).

[Fluent Bit](https://fluentbit.io/) collects logs.

The collected telemetry data are exposed so that you can view and analyze them with observability tools.

## Analyzing data

You can use the following in-cluster components to observe your applications' telemetry data:

- [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) receives and manages alerts coming from Prometheus. It can then forward the notifications about fired alerts to specific channels, such as Slack or an on-call paging system of your choice.
- [Grafana](https://grafana.com/docs/guides/getting_started/) provides a dashboard and a graph editor to visualize metrics collected from Prometheus.
- Kyma uses [Jaeger](https://www.jaegertracing.io/docs/) as a backend, which serves as the query mechanism for displaying information about traces.
- Kyma uses [Loki](https://github.com/grafana/loki), which is a lightweight Prometheus-like log management system.
- Kyma uses [Kiali](https://www.kiali.io) to enable validation, observe the Istio Service Mesh, and provide details on microservices included in the Service Mesh and connections between them.
