---
title: Kyma Eventing Metrics
---

Kyma Eventing provides several Grafana Dashboard with various [metrics](./evnt-02-eventing-metrics.md), so you can monitor statistics and other information in real time.

### Metrics Emitted by Eventing Publisher Proxy:

| Metric    |  Description |
|-------------|:--------------|
| **event_publish_to_messaging_server_errors_total** | Total number of errors while sending events to the messaging server. |
| **event_publish_to_messaging_server_latency** | Duration of sending events to the messaging server. |
| **event_type_published** | Total number of event publishing requests to the messaging server for a given event type. |
| **event_requests** | Total number of event publishing requests to the messaging server.  |

### Metrics Emitted by Eventing Controller:

| Metric    |  Description |
|-------------|:--------------|
| **event_type_subscribed** | All the eventTypes subscribed using the Subscription CRD. |
| **eventing_ec_delivery_per_subscription_total** | Number of dispatched events per subscription with information regarding the status code and its sink. |

### Metrics Emitted by NATS Exporter:

The [Prometheus NATS Exporter](https://github.com/nats-io/prometheus-nats-exporter) also emits metrics that you can monitor. Learn more about [NATS Monitoring](https://docs.nats.io/running-a-nats-service/configuration/monitoring#jetstream-information).  
 
