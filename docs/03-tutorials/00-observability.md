---
title: Observability
---

## Purpose

The following instructions describe the complete monitoring flow for your services in Kyma. You get the gist of monitoring applications, such as Prometheus, Grafana, and Alertmanager. You learn how and where you can observe and visualize your service metrics to monitor them for any alerting values.

Kyma comes with a Prometheus stack ([deprecated](https://github.com/kyma-project/website/blob/main/content/blog-posts/2022-12-09-monitoring-deprecation/index.md)), which is designed and sized to monitor Kyma's system components. We recommend to set up an additional Prometheus stack to monitor your custom metrics.

All the tutorials use the [`monitoring-custom-metrics`](https://github.com/kyma-project/examples/tree/main/prometheus/monitoring-custom-metrics) example and one of its services called `sample-metrics-8081`. This service exposes the `cpu_temperature_celsius` custom metric on the `/metrics` endpoint. This custom metric is the central element of the whole tutorial set. The metric value simulates the current processor temperature and changes randomly from 60 to 90 degrees Celsius. The alerting threshold in these tutorials is 75 degrees Celsius. If the temperature exceeds this value, the Grafana dashboard, PrometheusRule, and Alertmanager notifications you create inform you about this.

## Sequence of tasks

The instructions cover the following tasks:

 ![Monitoring tutorials](./assets/monitoring-tutorials.svg)

1. [**Deploy a custom Prometheus stack**](https://github.com/kyma-project/examples/blob/main/prometheus/README.md), in which you deploy the [kube-prometheus-stack](https://github.com/prometheus-operator/kube-prometheus) from the upstream Helm chart.

2. [**Observe application metrics**](https://github.com/kyma-project/examples/blob/main/prometheus/monitoring-custom-metrics/README.md), in which you redirect the `cpu_temperature_celsius` metric to the localhost and the Prometheus UI. You later observe how the metric value changes in the predefined 10 seconds interval in which Prometheus scrapes the metric values from the service's `/metrics` endpoint.

3. [**Create a Grafana dashboard**](https://github.com/kyma-project/examples/blob/main/prometheus/monitoring-grafana-dashboard/README.md), in which you create a Grafana dashboard of a Gauge type for the `cpu_temperature_celsius` metric. This dashboard shows explicitly when the CPU temperature is equal to or higher than the predefined threshold of 75 degrees Celsius, at which point the dashboard turns red.

4. [**Define alerting rules**](https://github.com/kyma-project/examples/blob/main/prometheus/monitoring-alert-rules/README.md), in which you define the `CPUTempHigh` alerting rule by creating a PrometheusRule resource. Prometheus accesses the `/metrics` endpoint every 10 seconds and validates the current value of the `cpu_temperature_celsius` metric. If the value is equal to or higher than 75 degrees Celsius, Prometheus waits for 10 seconds to recheck it. If the value still exceeds the threshold, Prometheus triggers the rule. You can observe both the rule and the alert it generates on the Prometheus dashboard.
