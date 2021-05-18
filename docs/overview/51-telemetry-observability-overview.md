---
title: Telemetry and Observability
type: Observability
---

Out of the box, Kyma provides tools to collect and expose **telemetry** data, such as metrics, traces, and log data.

Of course, you'll want to view and analyse the data you're collecting. This is where **observability** comes in. 

## Collecting data

The [OpenTelemetry](https://opentelemetry.io/) observability framework is the core tool where all of Kyma's raw data comes together. 

Among the data sources that flow into the OpenTelemetry collector are Kubernetes and Istio.

The collected telemetry data are exposed so that you can view and analyse them with the observability tools of your choice.

## Analysing data

Kyma supports a set of tools for in-cluster observability. 
We recommend that you also implement an observability solution of your choice outside your cluster, which has the advantage that you can use the data for troubleshooting and root cause analysis while your cluster is down (also, it doesn't eat into your applications' bandwith). 

### In-cluster observability

You can use the following in-cluster components to observe your applications' telemetry data:

- Prometheus
  [Prometheus](https://prometheus.io/docs/introduction) collects metrics from Pods. Metrics are the time-stamped data that provide information on the running jobs, workload, CPU consumption, memory usage, and more.
- Alertmanager
  [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) receives and manages alerts coming from Prometheus. It can then forward the notifications about fired alerts to specific channels, such as Slack or an on-call paging system of your choice.
- Grafana
  [Grafana](https://grafana.com/docs/guides/getting_started/) provides a dashboard and a graph editor to visualize metrics collected from Prometheus.
- Jaeger
  Kyma uses [Jaeger](https://www.jaegertracing.io/docs/) as a backend, which serves as the query mechanism for displaying information about traces.
- Loki
  Kyma uses [Loki](https://github.com/grafana/loki), which is a lightweight Prometheus-like log management system. Currently, Kyma supports the [Fluent Bit](https://fluentbit.io/) log collector.
- Kiali
  Kyma uses [Kiali](https://www.kiali.io) to enable validation, observe the Istio Service Mesh, and provide details on microservices included in the Service Mesh and connections between them.

See how to configure them for your needs under [Configuring In-Cluster Observability](link_tbd).

However, if your cluster is down, these components are down as well. This is why we recommend that you implement an observability solution outside your cluster.

All metrics relevant for observing the in-cluster Istio Service Mesh are collected separately. You can find more information about it on [Istio monitoring](link-to-istio-monitoring).
### External observability

If you want to use other observability tools than the ones within the cluster provided by Kyma out-of-the-box, you can easily do that, too. 
OpenTelemetry exposes the collected data *in the following way*:

  *TBD - input needed*

You can integrate various frameworks and libraries, such as:

  *TBD - input needed*

- *option 1*
- *option 2*
- *option 3*

Learn how to integrate Kyma with your preferred observability solution under [this tutorial](link-to-topic).

## Learn more

Interested in the architecture details? Check out the [logging architecture](arch-logging) and the [end-to-end monitoring flow](arch-monitoring).
