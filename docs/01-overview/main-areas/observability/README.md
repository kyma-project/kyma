---
title: What is Observability in Kyma?
---

Modern cloud applications consist of multiple components that can be scaled and deployed individually, and are focused on one concern using the most fitting technology - essentially, they're distributed systems. Monitoring such an application requires re-thinking: It's no longer an action you apply to a specific component, instead applications must actively expose insights! Monitoring becomes a federated aspect for all components, where every single component must actively expose its internal state.

Thus, the term "Observability" can be defined as a measure of how well internal states can be inferred from the application's external outputs. The insights that are  exposed are called "telemetry" or "signals" - usually metrics, traces and logs. They can be exposed by employing modern instrumentation.

Out of the box, Kyma provides tools to collect and expose **telemetry** data, such as metrics, traces, and log data. Of course, you'll want to view and analyze the data you're collecting. This is where **observability** tools come in.

## Collecting data

Kyma collects telemetry data with several in-cluster components:

- [Prometheus](https://prometheus.io/docs/introduction) collects metrics from Pods. Metrics are the time-stamped data that provide information on the running jobs, workload, CPU consumption, memory usage, and more. All metrics relevant for observing the in-cluster Istio Service Mesh are collected separately.

- [Fluent Bit](https://fluentbit.io/) collects logs.

- Traces are sent to [Jaeger](https://www.jaegertracing.io/docs).

The collected telemetry data are exposed so that you can view and analyze them with observability tools.

## Analyzing data

You can use the following in-cluster components to observe your applications' telemetry data:

- [Prometheus](https://prometheus.io/docs/introduction) is a lightweight backend for metrics.
- Kyma uses [Loki](https://github.com/grafana/loki), which is a lightweight Prometheus-like backend for logs.
- Kyma uses [Jaeger](https://www.jaegertracing.io/docs/) as a backend, which serves as the query mechanism for displaying information about traces.

- [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) receives and manages alerts coming from Prometheus. It can then forward the notifications about fired alerts to specific channels, such as Slack or an on-call paging system of your choice.
- [Grafana](https://grafana.com/docs/guides/getting_started/) provides a dashboard and a query editor to visualize metrics and logs collected from Prometheus and Loki.
- Kyma uses [Kiali](https://www.kiali.io) to enable validation, observe the Istio Service Mesh, and provide details on microservices included in the Service Mesh and connections between them.
