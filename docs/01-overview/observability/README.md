---
title: What is Observability in Kyma?
---

Out of the box, Kyma provides tools to collect and ship telemetry data using the [Telemetry Module](../telemetry/README.md). Of course, you'll want to view and analyze the data you're collecting. This is where observability tools come in.

## Data collection

Kyma collects telemetry data with the following in-cluster components:

- [Fluent Bit](https://fluentbit.io/) collects logs, provided using the [Telemetry Module](../telemetry/README.md).

- An [OTel Collector](https://opentelemetry.io/docs/collector/) collects traces, provided using the [Telemetry Module](../telemetry/README.md).

The collected telemetry data are exposed so that you can view and analyze them with observability tools.

> **NOTE:** Kyma's [telemetry component](../telemetry/README.md) supports providing your own output configuration for your application's logs and traces. With this, you can connect your own observability systems inside or outside the Kyma cluster with the Kyma backend.

## Data analysis

You can use the following in-cluster components to observe your applications' telemetry data:

- [Prometheus](https://prometheus.io/docs/introduction), a lightweight backend for metrics.
> **NOTE:** The Prometheus integration has been [deprecated](https://blogs.sap.com/2022/12/09/deprecation-of-prometheus-grafana-based-monitoring-in-sap-btp-kyma-runtime/) and is planned to be removed.
- [Grafana](https://grafana.com/docs/guides/getting_started/) to provide a dashboard and a query editor to visualize metrics collected from Prometheus.
> **NOTE:** The Grafana integration has been [deprecated](https://blogs.sap.com/2022/12/09/deprecation-of-prometheus-grafana-based-monitoring-in-sap-btp-kyma-runtime/) and is planned to be removed.

# Monitoring

> **NOTE:** Prometheus and Grafana are [deprecated](https://blogs.sap.com/2022/12/09/deprecation-of-prometheus-grafana-based-monitoring-in-sap-btp-kyma-runtime/) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

## Overview

For in-cluster monitoring, Kyma uses [Prometheus](https://prometheus.io/) as the open source monitoring and alerting toolkit that collects and stores metrics data. This data is consumed by several addons, including [Grafana](https://grafana.com/) for analytics and monitoring, and [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) for handling alerts.

Monitoring in Kyma is configured to collect all metrics relevant for observing the in-cluster [Istio](https://istio.io/latest/docs/concepts/observability/) Service Mesh. For diagrams of the default setup and the monitoring flow including Istio, see [Monitoring Architecture](../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md).

Learn how to [enable Grafana visualization](../../04-operation-guides/operations/obsv-03-enable-grafana-for-istio.md) and [enable mTLS for custom metrics](../../04-operation-guides/operations/obsv-04-enable-mtls-for-custom-metrics.md).

## Limitations

In the production profile, Prometheus stores up to **15 GB** of data for a maximum period of **30 days**. If the default size or time is exceeded, the oldest records are removed first. The evaluation profile has lower limits. For more information about profiles, see [Install Kyma: Choose resource consumption](../../04-operation-guides/operations/02-install-kyma.md#choose-resource-consumption).

The configured memory limits of the Prometheus and Prometheus-Istio instances define the number of time series samples that can be ingested.

The default resource configuration of the monitoring component in the production profile is sufficient to serve **800K time series in the Prometheus Pod**, and **400K time series in the Prometheus-Istio Pod**. The samples are deleted after 30 days or when reaching the storage limit of 15 GB.

The amount of generated time series in a Kyma cluster depends on the following factors:

* Number of Pods in the cluster
* Number of Nodes in the cluster
* Amount of exported (custom) metrics
* Label cardinality of metrics
* Number of buckets for histogram metrics
* Frequency of Pod recreation
* Topology of the Istio Service Mesh

You can see the number of ingested time series samples from the `prometheus_tsdb_head_series` metric, which is exported by the Prometheus itself. Furthermore, you can identify expensive metrics with the [TSDB Status](http://localhost:9090/tsdb-status) page.

# Telemetry

The page moved to the [Telemetry - Logs](https://kyma-project.io/#/telemetry-manager/user/02-logs) section.

# Useful links

If you're interested in learning more about the Observability area, check out these links:

- Learn how to set up the [Monitoring Flow](../../03-tutorials/00-observability.md) for your services in Kyma.

- Install a [custom Loki stack](https://github.com/kyma-project/examples/tree/main/loki).
- Install a [custom Jaeger stack](https://github.com/kyma-project/examples/tree/main/jaeger).
- Install a [custom Prometheus stack](https://github.com/kyma-project/examples/tree/main/prometheus).

- To collect and ship workload metrics to an OTLP endpoint, see [Install an OTLP-based metrics collector](https://github.com/kyma-project/examples/tree/main/metrics-otlp).
- Learn how to [access and expose](../../04-operation-guides/security/sec-06-access-expose-grafana.md) the services Grafana, Jaeger, and Kiali.

- Troubleshoot Observability-related issues:
  - [Prometheus Istio Server keeps crashing](../../04-operation-guides/troubleshooting/observability/obsv-01-troubleshoot-prometheus-istio-server-crash-oom.md)
  - [Trace backend shows fewer traces than expected](../../04-operation-guides/troubleshooting/observability/obsv-02-troubleshoot-trace-backend-shows-few-traces.md)

- Understand the architecture of Kyma's [monitoring](../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md), [logging](https://kyma-project.io/#/telemetry-manager/user/02-logs), and [tracing](https://kyma-project.io/#/telemetry-manager/user/03-traces) components.

- Find the [configuration parameters for Monitoring](../../05-technical-reference/00-configuration-parameters/obsv-01-configpara-observability.md).

- [Deploy Kiali](https://github.com/kyma-project/examples/blob/main/kiali/README.md) to a Kyma cluster
