---
title: Observability charts
---

You find the configurable parameters in charts and subcharts in the respective `values.yaml` files. The `values.yaml` files are fully documented and provide details on the configurable parameters and their customization.

To override the configuration, redefine the default values in your custom `values.yaml` file and [deploy them](../../04-operation-guides/operations/03-change-kyma-config-values.md).

### Monitoring

[Monitoring `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/values.yaml) and subcharts:

- [Grafana `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/grafana/values.yaml)
- [Kube-state-metrics `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/kube-state-metrics/values.yaml)
- [Prometheus Istio `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/prometheus-istio/values.yaml)
- [Prometheus Node Exporter `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/prometheus-node-exporter/values.yaml)
- [Prometheus Pushgateway `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/prometheus-pushgateway/values.yaml)

### Logging

[Logging `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/logging/values.yaml) and subcharts:

- [Fluent Bit `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/fluent-bit/values.yaml)
- [Loki `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/loki/values.yaml)

### Tracing

[Tracing `values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/tracing/values.yaml)

### Kiali

> **NOTE:** Kiali is [deprecated](https://kyma-project.io/blog/kiali-deprecation) and is planned to be removed with Kyma release 2.11. If you want to use Kiali, follow the steps to deploy Kiali yourself from our [examples](https://github.com/kyma-project/examples/blob/main/kiali/README.md).

[Kiali `values.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/kiali/values.yaml)
