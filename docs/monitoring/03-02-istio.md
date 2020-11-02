---
title: Istio Monitoring
type: Details
---

The monitoring chart is pre-configured to collect all metrics relevant for observing the in-cluster [Istio](https://istio.io/latest/docs/concepts/observability/) service mesh, including the proxy-level, service-level and control-plane metrics.

The concept for collecting the [service-level](https://istio.io/latest/docs/concepts/observability/#service-level-metrics) metrics is build around the Istio Proxy (implemented by Envoy), which is counting decentralized all communication inside the service mesh. After scraping these high cardinality metrics from the envoys, still an aggregation on service level is required to get the final service related metrics.

Following the [Istio Observability Best Practice](https://istio.io/latest/docs/ops/best-practices/observability/) the scraping and aggregating of the service-level metrics is done in a dedicated prometheus instance. That instance has the smallest possible data retention time configured as the raw metrics being scraped from the Istio Proxies have a high cardinality and are not further required to be kept. Instead, the aggregated metrics will be then scraped by the main prometheus instance via the federate endpoint.

The istop-related instance is a Deployment with name 'monitoring-prometheus-istio-server'. It has a small data retention time which should not be changed, no PersistentVolume attached and it has a hardcoded configuration. Resources like a ServiceMonitor will never be picked up by that instance.

![Istio Monitoring](./assets/monitoring-istio.svg)
