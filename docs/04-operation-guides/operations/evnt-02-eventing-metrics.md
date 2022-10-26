---
title: Kyma Eventing Metrics
---

Kyma Eventing provides several Grafana Dashboard with various [metrics](./evnt-02-eventing-metrics.md), so you can monitor statistics and other information in real time.
The metrics follow the [Prometheus naming convention](https://prometheus.io/docs/practices/naming/).

### Metrics Emitted by Eventing Publisher Proxy:

| Metric                                                          | Description                                                                                   |
|-----------------------------------------------------------------|:----------------------------------------------------------------------------------------------|
| **eventing_epp_errors_total**                                   | Total number of errors while sending events to the messaging server.                          |
| **eventing_epp_messaging_server_latency_duration_milliseconds** | Duration of sending events to the messaging server in milliseconds.                           |
| **epp_event_type_published_total**                              | Total number of event publishing requests to the NATS messaging server for a given eventType. |
| **eventing_epp_requests_total**                                 | Total number of event publishing requests to the messaging server.                            |

### Metrics Emitted by Eventing Controller:

| Metric                                      | Description                                                                    |
|---------------------------------------------|:-------------------------------------------------------------------------------|
| **nats_ec_event_type_subscribed_total**     | Total number of all the eventTypes subscribed using the Subscription CRD.      |
| **nats_ec_delivery_per_subscription_total** | Total number of dispatched events per subscription, with status code and sink. |

### Metrics Emitted by NATS Exporter:

The [Prometheus NATS Exporter](https://github.com/nats-io/prometheus-nats-exporter) also emits metrics that you can monitor. Learn more about [NATS Monitoring](https://docs.nats.io/running-a-nats-service/configuration/monitoring#jetstream-information).  
 
