---
title: Monitoring
---

> **NOTE:** Prometheus and Grafana are [deprecated](https://github.com/kyma-project/website/blob/main/content/blog-posts/2022-12-09-monitoring-deprecation/index.md) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

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
<!-- markdown-link-check-disable-next-line -->
You can see the number of ingested time series samples from the `prometheus_tsdb_head_series` metric, which is exported by the Prometheus itself. Furthermore, you can identify expensive metrics with the [TSDB Status](http://localhost:9090/tsdb-status) page.
