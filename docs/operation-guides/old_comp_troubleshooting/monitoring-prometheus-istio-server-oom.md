---
title: Prometheus Istio Server restarting or in crashback loop
type: Troubleshooting
---

Prometheus Istio Server scrapes metrics from all envoy side cars. It might crash because of OOM when the cardinality of the Istio metrics increases too much. This usually happens when a high amount of workloads perform a lot of communication to other workloads. This issue can be tackled by following these steps:

1. Increase the memory for `prometheus-istio-server`:
    ```bash
    kubect edit deployment -n kyma monitoring-prometheus-istio-server
    ```
    Increase the limits for memory:
    ```yaml
    resources:
      limits:
        cpu: 600m
        memory: 2000Mi
      requests:
        cpu: 40m
        memory: 200Mi
    ```
2. Drop labels for the Istio metrics. 
    
   Edit the ConfigMap for `prometheus-istio server`: 
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

## Create an override
Follow these steps to [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation) the existing configuration

1. Add and apply a ConfigMap in the `kyma-installer` Namespace in which you set the value for the **memory limit** attribute to 4Gi and/or **drop the labels** from Istio metrics.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: monitoring-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: monitoring
    kyma-project.io/installation: ""
data:
  prometheus-istio.envoyStats.labeldropRegex: "^(grpc_response_status|source_version|source_principal|source_app|response_flags|request_protocol|destination_version|destination_principal|destination_app|destination_canonical_service|destination_canonical_revision|source_canonical_revision|source_canonical_service)$"
  prometheus-istio.server.resources.limits.memory: "4Gi"
EOF
```
> **CAUTION:** The side effect of the change to `prometheus-istio.envoyStats.labeldropRegex` (to drop additional labels ) is graphs in Kiali will not work.
