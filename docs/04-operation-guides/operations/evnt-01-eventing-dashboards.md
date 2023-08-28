---
title: Grafana Dashboards for Kyma Eventing
---

Kyma Eventing provides several Grafana Dashboard with various [metrics](./evnt-02-eventing-metrics.md), so you can monitor statistics and other information in real time.

> **NOTE:** See how to [expose Grafana securely](../security/sec-06-access-expose-grafana.md) for easier access in the future.

> **NOTE:** Prometheus and Grafana are [deprecated](https://github.com/kyma-project/website/blob/main/content/blog-posts/2022-12-09-monitoring-deprecation/index.md) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

1. To access the Grafana Dashboard, port-forward the Grafana Service to localhost:

   ```bash
   kubectl -n kyma-system port-forward svc/monitoring-grafana 8081:80
   ```

2. Access the Grafana Dashboard on [localhost:8081](http://localhost:8081).

3. Select the Grafana Dashboard with the desired information about Kyma Eventing:

| Dashboard    |  Description |
|-------------|:--------------|
| **Eventing Pods** | Information about CPU, memory, and network resources consumed by the Kyma Eventing Pods. |
| **Eventing Delivery** | Statistics of HTTP requests to event publisher proxy and NATS subscribers. Also contains successful and failed events published, as well as delivery statistics and analytical information for published and subscribed event types and their respective mapping. |
| **Eventing Latency** | Latency information in the event delivery lifecycle from Eventing publisher proxy to Eventing backend servers, and from dispatcher to subscriber.  |
| **NATS Servers** | Information about CPU, memory, and network resources consumed by the NATS Servers  |
| **NATS JetStream** | NATS JetStream-specific information about streams and consumers.  |
