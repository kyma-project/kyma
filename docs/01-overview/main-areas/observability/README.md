---
title: What is Observability in Kyma?
---

Out of the box, Kyma provides tools to collect and ship telemetry data using the [Telemetry Module](../telemetry/README.md). Of course, you'll want to view and analyze the data you're collecting. This is where observability tools come in.

## Data Collection

Kyma collects telemetry data with the following in-cluster components:

- [Fluent Bit](https://fluentbit.io/) collects logs, provided using the [Telemetry Module](../telemetry/README.md).

- An [OTel Collector](https://opentelemetry.io/docs/collector/) collects traces, provided using the [Telemetry Module](../telemetry/README.md).

The collected telemetry data are exposed so that you can view and analyze them with observability tools.

> **NOTE:** Kyma's [telemetry component](../telemetry/README.md) supports providing your own output configuration for your application's logs and traces. With this, you can connect your own observability systems inside or outside the Kyma cluster with the Kyma backend.

## Data analysis

You can use the following in-cluster components to observe your applications' telemetry data:

- [Prometheus](https://prometheus.io/docs/introduction), a lightweight backend for metrics.
> **NOTE:** The Prometheus integration has been [deprecated](https://kyma-project.io/blog/2022/12/9/monitoring-deprecation) and is planned to be removed.
- [Loki](https://grafana.com/oss/loki/), a lightweight backend for metrics. 
> **NOTE:** The Loki integration has been [deprecated](https://kyma-project.io/blog/2022/11/2/loki-deprecation/) and is planned to be removed.
- [Grafana](https://grafana.com/docs/guides/getting_started/) to provide a dashboard and a query editor to visualize metrics and logs collected from Prometheus and Loki.
> **NOTE:** The Grafana integration has been [deprecated](https://kyma-project.io/blog/2022/12/9/monitoring-deprecation) and is planned to be removed.
