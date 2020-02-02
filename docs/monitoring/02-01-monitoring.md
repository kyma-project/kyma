---
title: Architecture
---

Before you learn how the complete metric flow looks in Kyma, read about components and resources that are crucial elements of the monitoring flow in Kyma.

![](./assets/monitoring-architecture.svg)


1. [**Prometheus Operator**](https://github.com/coreos/prometheus-operator) creates a Prometheus instance, manages its deployment, and provides configuration for it. It also operates ServiceMonitor custom resources that specify monitoring definitions for groups of services. Prometheus Operator is a prerequisite for installing other core monitoring components, such as Alertmanager, Node Exporter and Grafana. 

2. [**Prometheus**](https://prometheus.io/docs/introduction)  collects metrics from Pods. The metrics are the time-stamped data that provide information on the running jobs, workload, CPU consumption, memory usage, and more. These come from  [**kube-state-metrics**](https://github.com/kubernetes/kube-state-metrics) which is a simple service responsible for generating metrics for the objects, such as Pods or Nodes. Pods can also contain applications with custom metrics, such as the total storage space available in the MinIO server. Prometheus stores this polled data in a time-series database (TSDB) and runs rules over them to generate alerts if it detects any metric anomalies. Prometheus uses the[**Node Exporter**](https://github.com/mindprince/nvidia_gpu_prometheus_exporter)  export existing metrics from external systems as Prometheus metrics. 

3. **ServiceMonitors** monitor services and specify the endpoints from which Prometheus should poll the metrics. Even if you expose a handful of metrics in your application, Prometheus polls only those from the `/metrics` endpoints of ports specified in ServiceMonitor CRDs. 

3. [**Alertmanager**](https://prometheus.io/docs/alerting/alertmanager/) receives alerts from Prometheus and forwards this data to configured Slack or Victor Ops channels.

  > **NOTE:** There are no notification channels configured in the default monitoring installation. The current configuration allows you to add either Slack or Victor Ops channels.


6. [**Grafana**](https://grafana.com/docs/guides/getting_started/) that provides a dashboard and a graph editor to visualize metrics collected from the Prometheus API. Grafana uses the query language called [PromQL](https://prometheus.io/docs/prometheus/latest/querying/basics/) to select and aggregate metrics data from the Prometheus database. To access the Grafana UI, use the `https://grafana.{DOMAIN}` address, where `{DOMAIN}` is the domain of your Kyma cluster.

## Related resources

Monitoring in Kyma also relies heavily on these custom resources:

- **PrometheusRules** define alert conditions for metrics. They are configured in Prometheus as PrometheusRule custom resource definitions (CRDs). Kyma provides a set of out-of-the-box alerting rules that are passed from Prometheus to Alertmanager. The definitions of such rules specify the alert logic, the value at which alerts are triggered, the alerts' severity, and more. If you pre-define specific Slack or Victor Ops channels, Alertmanager displays the alerts in the channel each time the alerts are triggered.



## End-to-end monitoring flow

The complete monitoring flow in Kyma comes down to these components and steps:

![](./assets/monitoring-flow.svg)

1. Upon Kyma installation on a cluster, Prometheus Operator creates a Prometheus instance with the default configuration.
2. The Prometheus server periodically polls all metrics exposed on `/metrics` endpoints of ports specified in ServiceMonitor CRDs. Prometheus stores these metrics in a time-series database.
3. If Prometheus detects any metric values matching the logic of alerting rules, it triggers the alerts and passes them to Alertmanager.
4. If you manually configure a notification channel, you can instantly receive detailed information on metric alerts detected by Prometheus.
5. You can visualize metrics and track their historical data on Grafana dashboards.
