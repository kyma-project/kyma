---
title: Enable Grafana for Istio monitoring    
---

By default, `monitoring-prometheus-istio-server` is not provided as a data source in Grafana.

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
