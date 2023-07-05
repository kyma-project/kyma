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

    > **NOTE:** Prometheus and Grafana are [deprecated](https://kyma-project.io/blog/2022/12/9/monitoring-deprecation) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).
