---
title: Monitoring Kyma Eventing
---

Kyma Eventing provides several Grafana Dashboards, providing statistics and other information for real-time monitoring of the Eventing backbone.

To access the Grafana Dashboard, we will port-forward the Grafana Service to localhost:
```bash
kubectl -n kyma-system port-forward svc/monitoring-grafana 8081:80
```

Below is the list of Grafana Dashboards provided by Kyma Eventing:

| Dashboard    |  Description |
|-------------|:--------------|
| **Eventing Delivery** | Shows the statistics of HTTP requests to application validator, event publisher proxy and NATS subscribers. |
| **Eventing Delivery per Subscription** | Shows the successful and failed event delivery statistics per subscription. |
| **Eventing Latency** | Shows the latency information in the Event delivery lifecycle from eventing publisher proxy to eventing backend servers and dispatcher to subscriber.  |
| **NATS Servers** | Shows the information about compute, memory and network resources consumed by the NATS Servers  |
| **NATS JetStream** | Shows NATS JetStream specific information about Streams and Consumers.  |
| **NATS JetStream Event Types Summary** | Shows analytical information for published and subscribed Event Types and their respective mapping. |
