---
title: Monitoring in Kyma
---

[Prometheus](https://prometheus.io/) is the open source monitoring and alerting toolkit that provides the telemetry data. This data is consumed by different addons, including [Grafana](https://grafana.com/) for analytics and monitoring, and [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) for handling alerts. 

By default, Prometheus stores up to 15 GB of data for a maximum period of 30 days. If the default size or time is exceeded, the oldest records are removed first.

Monitoring in Kyma is configured to collect all metrics relevant for observing the in-cluster [Istio](https://istio.io/latest/docs/concepts/observability/) Service Mesh. For diagrams of the default setup and the monitoring flow with Istio, see [Istio Monitoring Architecture](../../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring-istio.md). Learn how to [enable Grafana visualization](../../../04-operation-guides/operations/obsv-03-enable-grafana-for-istio.md) and [enable mTLS for custom metrics](../../../04-operation-guides/operations/obsv-04-enable-mtls-istio). 
