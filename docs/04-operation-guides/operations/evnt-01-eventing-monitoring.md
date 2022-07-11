---
title: Monitoring Kyma Eventing
---

Kyma Eventing provides several Grafana Dashboards so you can monitor statistics and other information for the Eventing backbone in real time.

1. To access the Grafana Dashboard, port-forward the Grafana Service to localhost:

   ```bash
   kubectl -n kyma-system port-forward svc/monitoring-grafana 8081:80
   ```

2. Access the Grafana Dashboard on [localhost:8081](http://localhost:8081).

3. Select the Grafana Dashboard with the desired information about Kyma Eventing:

| Dashboard    |  Description |
|-------------|:--------------|
| **Eventing Pods** | Information about CPU, memory, and network resources consumed by the Kyma Eventing Pods. |
| **Eventing Delivery** | Statistics of HTTP requests to application validator, event publisher proxy and NATS subscribers. |
| **Eventing Delivery per Subscription** | Successful and failed event delivery statistics per subscription. |
| **Eventing Latency** | Latency information in the event delivery lifecycle from Eventing publisher proxy to Eventing backend servers, and from dispatcher to subscriber.  |
| **NATS Servers** | Information about CPU, memory, and network resources consumed by the NATS Servers  |
| **NATS JetStream** | NATS JetStream-specific information about streams and consumers.  |
| **NATS JetStream Event Types Summary** | Analytical information for published and subscribed event types and their respective mapping. |
