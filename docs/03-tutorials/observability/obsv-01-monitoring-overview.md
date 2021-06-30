---
title: Configure the Monitoring Flow
---

## Purpose

The following instructions describe the complete monitoring flow for your services in Kyma. You get the gist of Kyma built-in monitoring applications, such as Prometheus, Grafana, and Alertmanager. You will learn how and where you can observe and visualize your service metrics to monitor them for any alerting values.

All the tutorials use the [`monitoring-custom-metrics`](https://github.com/kyma-project/examples/tree/master/monitoring-custom-metrics) example and one of its services called `sample-metrics-8081`. This service exposes the `cpu_temperature_celsius` custom metric on the `/metrics` endpoint. This custom metric is the central element of the whole tutorial set. The metric value simulates the current processor temperature and changes randomly from 60 to 90 degrees Celsius. The alerting threshold in these tutorials is 75 degrees Celsius. If the temperature exceeds this value, the Grafana dashboard, Prometheus rule, and Alertmanager notifications you create inform you about this.

## Sequence of tasks

The instructions cover the following tasks:

 ![Monitoring tutorials](./assets/monitoring-tutorials.svg)

1. [**Observe application metrics**](obsv-02-observe-application-metrics.md), in which you redirect the `cpu_temperature_celsius` metric to the localhost and the Prometheus UI. You later observe how the metric value changes in the predefined 10 seconds interval in which Prometheus scrapes the metric values from the service's `/metrics` endpoint.

2. [**Create a Grafana dashboard**](obsv-03-create-and-configure-grafana-dashboard.md), in which you create a Grafana dashboard of a Gauge type for the `cpu_temperature_celsius` metric. This dashboard shows explicitly when the CPU temperature is equal to or higher than the predefined threshold of 75 degrees Celsius, at which point the dashboard turns red.

3. [**Define alerting rules**](obsv-04-define-alerting-rules-monitor.md), in which you define the `CPUTempHigh` alerting rule by creating a PrometheusRule resource. Prometheus accesses the `/metrics` endpoint every 10 seconds and validates the current value of the `cpu_temperature_celsius` metric. If the value is equal to or higher than 75 degrees Celsius, Prometheus waits for 10 seconds to recheck it. If the value still exceeds the threshold, Prometheus triggers the rule. You can observe both the rule and the alert it generates on the Prometheus dashboard.

4. [**Send notifications to Slack**](obsv-05-send-notifications.md), in which you configure Alertmanager to send notifications on Prometheus alerts to a Slack channel. This way, whenever Prometheus triggers or resolves the `CPUTempHigh` alert, Alertmanager sends a notification to the `test-monitoring-alerts` Slack channel defined for the tutorial.

5. [**Clean up the configuration**](obsv-06-clean-up-configuration.md), in which you remove the deployed example and all its resources from the cluster.
