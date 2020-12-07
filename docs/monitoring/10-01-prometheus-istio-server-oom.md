---
title: Prometheus Istio Server restarting or in crashback loop due to OOM
type: Troubleshooting
---

Prometheus Istio Server scrapes metrics from all the envoy side car. It might crash because of OOM when the cardinality of the istio metrics is increasing too much. This is usually is happening for a high amount of workloads performing a lot of communication to other workloads. This issue can be tackled in following ways:

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

## Create an override
Follow these steps to [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation) the existing configuration with a customized control plane definition.

1. Add and apply a ConfigMap in the `kyma-installer` Namespace in which you set the value for the **memory limit** attribute to 4Gi and **dropping the labels** from istio metrics.

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