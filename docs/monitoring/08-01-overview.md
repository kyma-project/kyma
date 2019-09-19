---
title: Overview
type: Tutorials
---

The set of monitoring tutorials you are about to read describes the complete monitoring flow for your metrics in Kyma. Going through the tutorials, you will get the gist of Kyma in-built monitoring applications, such as Prometheus, Grafana, and Alertmanager. You will also understand how and where you can observe and visualize your service metrics to monitor them for any alerting values.

All tutorials use the [`monitoring-custom-metrics`](https://github.com/kyma-project/examples/tree/master/monitoring-custom-metrics) example and one of its services called `sample-metrics-8081`. This service exposes the `cpu_temperature_celsius` custom metric on the `/metrics` endpoint. This custom metric is the central element of all tutorials. Its value aims to simulate the current processor temperature and changes randomly from 60 to 90 degrees Celsius. The alerting threshold in these tutorials is 75 degrees Celsius. If the temperature exceeds this value, the Grafana dashboard you create, and the Prometheus rule you set for monitoring the metric, clearly inform you of the alerting value.

The tutorial set contains:

1. [**Observe application metrics**](#tutorials-observe-application-metrics) in which you redirect the `cpu_temperature_celsius` metric to a localhost and the Prometheus UI. You later observe how the metric value changes in the predefined 10s interval in which Prometheus scrapes the metric values from the service `/metrics` endpoint.

2. [**Create a Grafana dashboard**](#tutorials-create-a-grafana-dashboard) in which you create a Grafana dashboard of a Gauge type for the `cpu_temperature_celsius` metric. This dashboard shows explicitly when the CPU temperature exceeds the predefined threshold of 75 degrees Celsius and turns red whenever this value is higher.

3. [**Define alerting rules**](#tutorials-define-alerting-rules) in which you define the `CPUTempHigh` alerting rule by creating a PrometheusRule resource. Prometheus accesses the `/metrics` endpoint every 10 seconds and validates the current value of the `cpu_temperature_celsius` metric. If the value is higher than 75 degrees Celsius, Prometheus waits for 10 seconds and triggers the rule. You can access and observe the rule through the Prometheus dashboard.

4. [**Send notifications to Slack**](#tutorials-send-notifications-to-slack) in which you configure Alertmanager to send notifications on Prometheus alerts to a Slack channel. This way, whenever Prometheus triggers or resolves the `CPUTempHigh` alert, Alertmanager sends a notification to the `test-monitoring-alerts` Slack channel.

See the diagram for an overview of tutorials purpose and tools used in them:

 ![Monitoring tutorials](./assets/monitoring-tutorials.svg)
