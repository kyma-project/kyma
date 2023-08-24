---
title: Enable Grafana for Istio monitoring    
---

By default, `monitoring-prometheus-istio-server` is not provided as a data source in Grafana.

> **NOTE:** Prometheus and Grafana are [deprecated](https://github.com/kyma-project/website/blob/main/content/blog-posts/2022-12-09-monitoring-deprecation/index.md) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

1. To enable `monitoring-prometheus-istio-server` as a data source in Grafana, provide a YAML file with the following values:

  ```yaml
  monitoring  
    prometheus-istio:
      grafana:
        datasource:
          enabled: "true"
  ```

2. Deploy the `values.yaml` file (see [Change Kyma settings](./03-change-kyma-config-values.md)).

3. Restart the Grafana deployment with the following command:

  ```bash
  kubectl rollout restart -n kyma-system deployment monitoring-grafana
  ```
