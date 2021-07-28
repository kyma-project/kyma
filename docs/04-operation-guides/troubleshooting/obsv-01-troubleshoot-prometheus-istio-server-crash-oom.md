---
title: Prometheus Istio Server restarting or in crashback loop
---

*work in progress*

## Condition

Prometheus Istio Server is restarting or in a crashback loop.

## Cause

Prometheus Istio Server scrapes metrics from all envoy side cars.

Because Istio telemetry V2 currently doesn't support the concept of metric expiry, the cardinality of the Istio metrics increases too much. This causes the container to be killed because of OOM.

For example, this may happen when a high number of workloads perform a lot of communication to other workloads, or when workloads are created and deleted dynamically.

There may be other causes for the Prometheus Istio Server to restart or crash, but the following steps only explain how to fix the OOM issue.

## Remedy

There are two ways to fix this:

- You can fix it on the Helm side
- You can configure your Kyma cluster with a specific values YAML file.

### Change the Helm configuration

1. Increase the memory for `prometheus-istio-server`:

    ```bash
    kubect edit deployment -n kyma monitoring-prometheus-istio-server

    ```

2. In your deployment resource, increase the limits for memory:

    ```yaml
    resources:
      limits:
        cpu: 600m
        memory: 2000Mi
      requests:
        cpu: 40m
        memory: 200Mi
    ```

3. Drop labels for the Istio metrics.

   Edit the values for `prometheus-istio server`:

    ```bash
    kubectl edit configmap -n kyma-system monitoring-prometheus-istio-server
    ```

    Edit:

    ```yaml
    metric_relabel_configs:
      - separator: ;
        regex: ^(grpc_response_status|source_version|destination_version|source_app|destination_app)$
        replacement: $1
        action: labeldrop
    ```

    Change regex to:

    ```yaml
    regex: ^(grpc_response_status|source_version|source_principal|source_app|response_flags|request_protocol|destination_version|destination_principal|destination_app|destination_canonical_service|destination_canonical_revision|source_canonical_revision|source_canonical_service)$
    ```

    Save the ConfigMap and restart `prometheus-istio-server` for the changes to take effect:

    ```bash
    kubectl rollout restart deployment -n kyma-system monitoring-prometheus-istio-server
    ```

    > **CAUTION:** The side effect of this change is graphs in Kiali will not work.

### Change the Kyma configuration

[Change the configuration](../../.../04-operation-guides/operations/03-change-kyma-config-values.md) with a values YAML file. You can set the value for the **memory limit** attribute to 4Gi, and/or **drop the labels** from Istio metrics.

> **CAUTION:** If you drop additional labels with `prometheus-istio.envoyStats.labeldropRegex`, graphs in Kiali will not work.

```yaml
monitoring
  prometheus-istio:
    envoyStats:
      labeldropRegex: "^(grpc_response_status|source_version|source_principal|source_app|response_flags|request_protocol|destination_version|destination_principal|destination_app|destination_canonical_service|destination_canonical_revision|source_canonical_revision|source_canonical_service)$"
    server:
      resources:
        limits:
          memory: "4Gi"
```
