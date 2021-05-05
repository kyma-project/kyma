---
title: Telemetry and Observability
type: Observability
---

Out of the box, Kyma provides **telemetry** tools to collect and expose raw data, such as metrics, traces, and log data.

Of course, you'll want to view and analyse the data you're collecting. This is where **observability** comes in. 

## Telemetry - collecting data

The [OpenTelemetry](https://opentelemetry.io/) observability framework is the core tool where all of Kyma's raw data comes together. 

Among the data sources that flow into the OpenTelemetry collector are Kubernetes and Istio. They are also collected in the Prometheus UI, where you can view them under **Targets**.

The collected telemetry data are exposed so that you can view and analyse them with the observability tools of your choice.

## Observability - analysing data

Kyma supports a set of tools for in-cluster observability. 
We recommend that you also implement an observability solution of your choice outside your cluster, which has the advantage that you can use the data for troubleshooting and root cause analysis while your cluster is down (also, it doesn't eat into your applications' bandwith). 

### In-cluster observability

You can use the following in-cluster components to observe your applications' telemetry data:

- [Prometheus](con-tbd)
- [Grafana](con-tbd)
- [Jaeger](con-tracing)
- [Loki](con-logging)

See how to configure them for your needs under [Configuring In-Cluster Observability](link_tbd.

However, if your cluster is down, these components are down as well. This is why we recommend that you implement an observability solution outside your cluster.

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