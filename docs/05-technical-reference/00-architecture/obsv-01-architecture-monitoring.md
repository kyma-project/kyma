---
title: Monitoring Architecture
---

## End-to-end monitoring flow

The monitoring flow in Kyma comes down to the following components and steps:

![End-to-end monitoring flow](./assets/obsv-monitoring-flow.svg)

1. Upon Kyma installation on a cluster, **Prometheus Operator** creates a **Prometheus** and an **Alertmanager** instance with the default configuration.
2. The Prometheus server periodically polls all metrics exposed on `/metrics` endpoints of <!-- ports specified in ServiceMonitor CRDs --> Pods. Prometheus stores these metrics in a time-series database.
3. When Prometheus detects any metric values matching the logic of alerting rules, it triggers the alerts and passes them to **Alertmanager**.
4. If you have configured a notification channel, you can instantly receive detailed information on metric alerts detected by Prometheus.
5. You can visualize metrics and track their historical data on **Grafana** dashboards.

Learn how to [set up the monitoring flow](../../03-tutorials/00-observability/README.md).

## Monitoring components

The diagram presents monitoring components and the way they interact with one another.

![Monitoring components](./assets/obsv-monitoring-architecture.svg)

1. [**Prometheus Operator**](https://github.com/coreos/prometheus-operator) creates a **Prometheus** instance, manages its deployment, and provides configuration for it. It also deploys **Alertmanager** and manages **Service Monitor** custom resources that specify monitoring definitions for groups of services.

2. [**Prometheus**](https://prometheus.io/docs/introduction) collects metrics from Pods.

   - Metrics specific to applications (like orders processed per second) are polled from the application pods.
   - Metrics common for all applications are scraped from specialized pods like the **node-exporter**.

Prometheus stores this polled data in a time-series database (TSDB) and runs rules over them to create aggregated metrics or generate alerts if it detects any metric anomalies.

   >**NOTE:** Besides this main Prometheus instance, there is a second Prometheus instance running in the `kyma-system` Namespace. This second instance is responsible for collecting and aggregating [Istio Service Mesh metrics](../../01-overview/main-areas/service-mesh/smsh-01-details.md).

3. You can use **Prometheus Rules** to define alert conditions for metrics. Kyma provides a set of out-of-the-box alerting rules. The definitions of such rules specify the alert logic, the value at which alerts are triggered, the alerts' severity, and more.

4. **Service Monitors** monitor services and specify the endpoints from which Prometheus polls the metrics. Even if you expose a handful of metrics in your application, Prometheus polls only those from the `/metrics` endpoints of ports specified in Service Monitor CRDs.

5. [**Alertmanager**](https://prometheus.io/docs/alerting/alertmanager/) receives alerts from Prometheus and forwards this data to configured notification channels like Slack or Victor Ops.

    >**NOTE:** By default, no notification channels are configured. You need to [set them up](../../03-tutorials/00-observability/obsv-04-send-notifications.md).

6. [**Grafana**](https://grafana.com/docs/guides/getting_started/) provides a dashboard and a graph editor to visualize metrics collected from the Prometheus API. Grafana uses the query language [PromQL](https://prometheus.io/docs/prometheus/latest/querying/basics/) to select and aggregate metrics data from the Prometheus database. Learn how to [access the Grafana UI](../../04-operation-guides/security/sec-06-access-expose-kiali-grafana.md). 
