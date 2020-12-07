---
title: Prometheus Istio Server restarting or in crashback loop due to OOM
type: Troubleshooting
---

Prometheus Istio Server scrapes metricd from all the envoy side car. It can restart or have crashbackloop due to insufficient memory. This issue can be tackled in following ways:

1. Increase the memory for prometheus-istio server
    ```bash
    kubect edit deployment -n kyma monitoring-prometheus-istio-server
    ```
    Increase the limits for memory
    ```yaml
    resources:
      limits:
        cpu: 600m
        memory: 2000Mi
      requests:
        cpu: 40m
        memory: 200Mi
    ```
2. Drop labels for the istio metrics.
    
   Edit the configmap for prometheus-istio server
   ```bash
    kubectl edit configmap -n kyma-system monitoring-prometheus-istio-server
    ```
    Edit following
    ```yaml
    metric_relabel_configs:
      - separator: ;
        regex: ^(grpc_response_status|source_version|destination_version|source_app|destination_app)$
        replacement: $1
        action: labeldrop
    ```
    Change regex to
    ```yaml
    regex: ^(grpc_response_status|source_version|source_principal|source_app|response_flags|request_protocol|destination_version|destination_principal|destination_app|destination_canonical_service|destination_canonical_revision|source_canonical_revision|source_canonical_service)$
    ```
    save the configmap and restart prometheus-istio server for the config map changes to take effect
    ```bash
    kubectl rollout restart deployment -n kyma-system monitoring-prometheus-istio-server
    ```
    > Warning: The side effect of this change is graphs in kiali would not work.

