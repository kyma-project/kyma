---
title: Kyma Eventing Metrics
---

Kyma Eventing provides several Grafana Dashboard with various [metrics](./evnt-02-eventing-metrics.md), so you can monitor statistics and other information in real time.
The metrics follow the [Prometheus naming convention](https://prometheus.io/docs/practices/naming/).

> **NOTE:** Prometheus and Grafana are [deprecated](https://kyma-project.io/blog/2022/12/9/monitoring-deprecation) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

### Metrics Emitted by Eventing Publisher Proxy:

| Metric                                                          | Description                                                                                   |
|-----------------------------------------------------------------|:----------------------------------------------------------------------------------------------|
| **eventing_epp_backend_errors_total**                           | Total number of backend errors while sending events to the messaging server.                  |
| **eventing_epp_backend_duration_milliseconds**                  | Duration of sending events to the messaging server in milliseconds.                           |
| **eventing_epp_requests_duration_milliseconds**                 | Duration of processing an incoming request that includes sending the event to the backend.    |
| **eventing_epp_backend_requests_total**                         | Total number of event publishing requests to the NATS messaging server.                       |
| **eventing_epp_requests_total**                                 | Total number of publishing requests to the EPP.                                               |
| **eventing_epp_event_type_published_total**                     | Total number of event publishing requests to the NATS messaging server for a given eventType. |

### Metrics Emitted by Eventing Controller:

| Metric                                      | Description                                                                    |
|---------------------------------------------|:-------------------------------------------------------------------------------|
| **nats_ec_event_type_subscribed_total**     | Total number of all the eventTypes subscribed using the Subscription CRD.      |
| **nats_ec_delivery_per_subscription_total** | Total number of dispatched events per subscription, with status code and sink. |

### Metrics Emitted by NATS Exporter:

The [Prometheus NATS Exporter](https://github.com/nats-io/prometheus-nats-exporter) also emits metrics that you can monitor. Learn more about [NATS Monitoring](https://docs.nats.io/running-a-nats-service/configuration/monitoring#jetstream-information).  
 
