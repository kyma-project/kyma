---
title: Monitoring profiles
type: Configuration
---
## Overview

To ensure optimal performance and avoid high memory and CPU consumption, you can install Kyma with one of the Monitoring profiles. 

### Default profile 

The default profile is used in Kyma cluster installation when you deploy Kyma with Monitoring enabled. You can use it for development purposes but bear in mind that it is not production-ready. The profile defines short data retention time (1 day) which may not be enough to identify and solve issues in case of prolonged troubleshooting. To make Monitoring production-ready and avoid potential issues, configure Monitoring to [use the production profile](#configuration-monitoring-profiles-use-production-profile).

### Production profile

To make sure Monitoring runs in a production environment, this profile introduces the following changes: 

* Increased retention time to prevent data loss in case of prolonged troubleshooting 
* Increased memory and CPU values to ensure stable performance 

### Local profile

If you install Kyma locally on Minikube, Monitoring uses a lightweight configuration by default to avoid high memory and CPU consumption. 

## Parameters 

The table shows the parameters of each profile and their values:

 Parameter  | Description | Default profile| Production profile | Local profile|
|-----------|-------------|----------------|--------------------|--------------|
| **prometheus.prometheusSpec.retentionSize** | Maximum number of bytes that storage blocks can use. The oldest data will be removed first. | `2GB` | `15GB` | `256MB` |
| **prometheus.prometheusSpec.retention** | Time period for which Prometheus stores the metrics. |`1d`| `30d` | `2h`|
| **prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage** | Amount of storage requested by the Prometheus Pod. |`10Gi`| `20Gi` | `1Gi` |
| **prometheus.prometheusSpec.resources.limits.cpu** | Maximum number of CPUs available for the Prometheus Pod to use. | `600m`| `1` | `150m`|
| **prometheus.prometheusSpec.resources.limits.memory** | Maximum amount of memory available for the Prometheus Pod to use. |`2Gi` | `4Gi` |`800Mi`|
| **prometheus.prometheusSpec.resources.requests.cpu** |  Number of CPUs requested by the Prometheus Pod to operate.| `200m`| `300m` | `100m` |
| **prometheus.prometheusSpec.resources.requests.memory** | Amount of memory requested by the Prometheus Pod to operate. | `600Mi`| `1Gi` | `200Mi` |
| **alertmanager.alertmanagerSpec.retention** | Time period for which Alertmanager retains data.| `120h` | `240h` | `1h` |
| **grafana.persistence.enabled**| Parameter that enables storing Grafana database on a PersistentVolume |`true`|`true`|`false`|
| **prometheus-istio.server.resources.requests.memory** |  Maximum amount of memory available for the Prometheus-Istio Pod to use.| `200Mi`| `200Mi`|`200Mi`|
| **prometheus-istio.server.resources.limits.memory** |  Maximum amount of memory available for the Prometheus-Istio Pod to use.| `3Gi`| `4Gi`|`400Mi`|

## Use profiles

The default and local profiles are installed automatically during cluster and local installation respectively. The production profile is a Helm override you can apply before Kyma installation or in the runtime. 

### Production profile 

You can deploy a Kyma cluster with Monitoring configured to use the production profile, or add the configuration in the runtime. Follow these steps:

<div tabs>
  <details>
  <summary>
  Install Kyma with production-ready Monitoring
 </summary>

1. Create a Kubernetes cluster for Kyma installation.

2. Apply an override that forces Monitoring to use the production profile:

  ```bash
  cat <<EOF | kubectl apply -f -
  ---
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
    prometheus.prometheusSpec.retentionSize: "15GB"
    prometheus.prometheusSpec.retention: "30d"
    prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage: "20Gi"
    prometheus.prometheusSpec.resources.limits.cpu: "1"
    prometheus.prometheusSpec.resources.limits.memory: "4Gi"
    prometheus.prometheusSpec.resources.requests.cpu: "300m"
    prometheus.prometheusSpec.resources.requests.memory: "1Gi"
    prometheus-istio.server.resources.limits.memory: "4Gi"
    alertmanager.alertmanagerSpec.retention: "240h"
  EOF
  ```
  </details>
  <details>
  <summary>
  Enable configuration in a running cluster
  </summary>

  1. Apply an override that forces Monitoring to use the production profile:

  ```bash
    cat <<EOF | kubectl apply -f -
    ---
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
      prometheus.prometheusSpec.retentionSize: "15GB"
      prometheus.prometheusSpec.retention: "30d"
      prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage: "20Gi"
      prometheus.prometheusSpec.resources.limits.cpu: "1"
      prometheus.prometheusSpec.resources.limits.memory: "4Gi"
      prometheus.prometheusSpec.resources.requests.cpu: "300m"
      prometheus.prometheusSpec.resources.requests.memory: "1Gi"
      prometheus-istio.server.resources.limits.memory: "4Gi"
      alertmanager.alertmanagerSpec.retention: "240h"
    EOF
  ```
  2. Run the [cluster update process](/root/kyma/#installation-update-kyma).

When the production overrides are applied to an already installed Kyma cluster, then the changes to the storage size of the PVC for Prometheus will not be applied. This is because the underlying Cloud infrastructure might not support dynamic resizing of the PVC. For a workaround, follow these steps:  

>**CAUTION:** This workaround will delete existing metrics, as it creates a new persistent storage.

After the cluster update process is finished, proceed to apply the workaround:
1. Delete the Prometheus StatefulSet:
```bash
kubectl delete statefulset -n kyma-system  prometheus-monitoring-prometheus
```
2. Delete the PVC for the StatefulSet:
```bash
kubectl delete statefulset -n kyma-system prometheus-monitoring-prometheus-db-prometheus-monitoring-prometheus-0
```
After this, Prometheus operator should create a new PVC and a new StatefulSet. Verify if they are present.

3. Check if PVC has been successfully created:
```bash
kubect get pvc -n kyma-system prometheus-monitoring-prometheus-db-prometheus-monitoring-prometheus-0
```
Check the column `CAPACITY` and verify that `20Gi` is set as the new value.

4. Check if the StatefulSet has been created successfully:
```bash
kubectl get statefulsets.apps -n kyma-system prometheus-monitoring-prometheus
```
Check if the value in column `READY` is `1/1`.
  </details>
</div>
