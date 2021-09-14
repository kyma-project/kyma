---
title: Monitoring charts
---

You find the configurable parameters for Monitoring in the [Monitoring values.yaml](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/values.yaml) file and in the following subcharts:

- [Grafana](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/grafana/values.yaml)
- [Kube-state-metrics](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/kube-state-metrics/values.yaml)
- [Prometheus Istio](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/prometheus-istio/values.yaml)
- [Prometheus Node Exporter](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/prometheus-node-exporter/values.yaml)
- [Prometheus Pushgateway](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/charts/prometheus-pushgateway/values.yaml)

The `values.yaml` files are fully documented and provide details on the configurable parameters and their customization.

To override the configuration, modify the default values of the `values.yaml` file and [deploy them](../../04-operation-guides/operations/03-change-kyma-config-values.md).
