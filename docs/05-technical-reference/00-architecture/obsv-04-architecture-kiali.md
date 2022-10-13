---
title: Kiali Architecture
---

> **NOTE:** Kiali is [deprecated](https://kyma-project.io/blog/kiali-deprecation) and is planned to be removed with Kyma release 2.11. If you want to use Kiali, follow the steps to deploy Kiali yourself from our [examples](https://github.com/kyma-project/examples/blob/main/kiali/README.md).

The following diagram presents the overall Kiali architecture and the way the components interact with each other.

![Kiali architecture](./assets/obsv-kiali-architecture.svg)

Use the Kyma Console or direct URL to access Kiali. Learn more about [accessing Kiali](../../04-operation-guides/security/sec-06-access-expose-grafana.md).

Kiali collects the information on the cluster health from the following sources:

* Istio metrics scraped by Prometheus
* Kubernetes API server, which provides data on the cluster state
* Trace data collected by Jaeger
