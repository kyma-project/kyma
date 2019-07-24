---
title: Architecture
---

Before you learn how the complete metric flow looks like in Kyma, read about component and resources that are crucial elements that participate in the monitoring flow in Kyma.

## Components

The main monitoring components include:

- **Prometheus Operator** that creates a Prometheus instance, manages its deployment, and provides configuration for it. It also operates ServiceMonitor custom resources that specify the monitoring definitions for groups of services. The Prometheus Operator is a prerequisite for installing core monitoring components, such as Alertmanager and Grafana.

  For more details, [read](https://github.com/coreos/prometheus-operator) the official documentation.

- **Prometheus** that collects metrics from Pods. The metrics are the time-stamped data that provide information on jobs started, workload, CPU consumption, memory usage, and more. Pods can also contain applications with custom metrics, such as the total storage space available in the Minio server. Prometheus stores this polled data in a time-series database (TSDB) and runs rules over them to generate alerts if it detects any metric anomalies.

  For more details, [read](https://prometheus.io/docs/introduction) the official documentation.

- **Grafana** that provides a dashboard and graph editor to visualize metrics collected from the Prometheus API. Grafana uses the query language called [PromQL](https://prometheus.io/docs/prometheus/latest/querying/basics/) to select and aggregate the metrics data from the Prometheus database. To access the Grafana UI, use the `https://grafana.{DOMAIN}` address, where {DOMAIN} is the domain of your Kyma cluster.

  For more details, [read](https://grafana.com/docs/guides/getting_started/) the official documentation.

- **Alertmanager** that receives alerts from Prometheus and forwards this data to configured Slack or Victor Ops channels.

  > **NOTE:** There are no notification channels configured in the default monitoring installation. The current configuration allows you to add either Slack or Victor Ops channels.

  For more details, [read](https://prometheus.io/docs/alerting/alertmanager/) the official documentation.

## Related resources

Monitoring in Kyma also relies heavily on these custom resources:

- **Alert rules** define alert conditions for metrics. They are configured in Prometheus as PrometheusRule custom resource definitions (CRDs). Kyma provides a set of out-of-the-box alert rules that are passed from Prometheus to the Alertmanager. The definitions of such rules specify the alert logic, the value at which alerts are triggered, the alerts' severity, and more. If you pre-define specific Slack or Victor Ops channels, the Alertmanager fires the alerts onto the channel each time the alerts are triggered.

- **Service Monitors** are custom resource definitions that specify the endpoints from which Prometheus should poll the metrics. Even if you expose a handful of metrics in your application, Prometheus polls only those from the endpoints specified in a Service Monitor custom resource definitions.

## End-to-end monitoring flow

The complete monitoring flow in Kyma comes down to these components and steps:

![](./assets/monitoring-architecture.svg)

1. Upon Kyma installation on a cluster, Prometheus Operator creates a Prometheus instance with the default configuration.
2. The Prometheus server periodically polls all metrics exposed on endpoints of services specified in Service Monitors. Prometheus stores these metrics in a time-series database.
3. If Prometheus detects any anomalies in metrics that are covered by the alert rules, Prometheus fires them and passes the information to the Alertmanager.
4. If you manually configure a notification channel, you can instantly receive detailed information on the metric alerts detected by Prometheus.
5. You can visualize metrics and track their historical data on Grafana dashboards.
