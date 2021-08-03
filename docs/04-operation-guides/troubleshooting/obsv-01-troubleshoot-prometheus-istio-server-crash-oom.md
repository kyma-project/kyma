---
title: Prometheus Istio Server restarting or in crashback loop
---

## Condition

Prometheus Istio Server is restarting or in a crashback loop.

## Cause

Prometheus Istio Server scrapes metrics from all envoy side cars, which may lead to OOM issues.

For example, this can happen when a high number of workloads perform a lot of communication to other workloads, or when workloads are created and deleted dynamically.

In such cases, the cardinality of the Istio metrics may increase too much and cause the container to be killed because of OOM (Istio telemetry V2 currently doesn't support the concept of metric expiry).

There can be other causes for the Prometheus Istio Server to restart or crash, but the following istructions focus on preventing the OOM issue.

## Remedy

To prevent the OOM issue, you can increase the memory limit, and you can drop additional labels.

> **CAUTION:** Dropping additional labels with `prometheus-istio.envoyStats.labeldropRegex` has the side effect that graphs in Kiali will not work.

For both solution, you can choose to apply them either on the Helm side, or in your Kyma cluster configuration.

### Change the Helm configuration

1. To increase the memory for `prometheus-istio-server`, run the following command:

  ```bash
  kubect edit deployment -n kyma monitoring-prometheus-istio-server
  ```

2. In your deployment resource, set the following limits for memory:

  ```yaml
  resources:
    limits:
      cpu: 600m
      memory: 6000Mi
    requests:
      cpu: 40m
      memory: 200Mi
  ```

3. To drop labels for the Istio metrics, edit `prometheus-istio server`:

  ```bash
  kubectl edit configmap -n kyma-system monitoring-prometheus-istio-server
  ```

4. Edit the values:

  ```yaml
  metric_relabel_configs:
    - separator: ;
      regex: ^(grpc_response_status|source_version|destination_version|source_app|destination_app)$
      replacement: $1
      action: labeldrop
  ```

5. Change regex in the following way:

  ```yaml
  regex: ^(grpc_response_status|source_version|source_principal|source_app|response_flags|request_protocol|destination_version|destination_principal|destination_app|destination_canonical_service|destination_canonical_revision|source_canonical_revision|source_canonical_service)$
  ```

6. Save the ConfigMap and restart `prometheus-istio-server` for the changes to take effect:

  ```bash
  kubectl rollout restart deployment -n kyma-system monitoring-prometheus-istio-server
  ```

### Change the Kyma configuration

To [change the configuration](../../.../04-operation-guides/operations/03-change-kyma-config-values.md), deploy a values YAML file.

You can set the value for the **memory limit** attribute to 6Gi, and/or **drop the labels** from Istio metrics.

```yaml
monitoring:
  prometheus-istio:
    envoyStats:
      labeldropRegex: "^(grpc_response_status|source_version|source_principal|source_app|response_flags|request_protocol|destination_version|destination_principal|destination_app|destination_canonical_service|destination_canonical_revision|source_canonical_revision|source_canonical_service)$"
    server:
      resources:
        limits:
          memory: "6Gi"
```
