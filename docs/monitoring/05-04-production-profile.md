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
| **retentionSize** | Maximum number of bytes that storage blocks can use. The oldest data will be removed first. | `2GB` | `15GB` | `500MB` | 
| **retention** | Time period for which Prometheus stores metrics in an in-memory database. Prometheus stores the recent data for the specified amount of time to avoid reading all data from the disk. This parameter only applies to in-memory storage.|`1d`| `30d` | `2h`|
| **prometheusSpec.volumeClaimTemplate.spec.resources.requests.storage** | Amount of storage requested by the Prometheus Pod. |`10Gi`| `20Gi` | `1Gi` |
| **prometheusSpec.resources.limits.cpu** | Maximum number of CPUs available for the Prometheus Pod to use. | `600m`| `600m` | `300m`|
| **prometheusSpec.resources.limits.memory** | Maximum amount of memory available for the Prometheus Pod to use. |`1500Mi` | `2Gi` |`250Mi`|
| **prometheusSpec.resources.requests.cpu** |  Number of CPUs requested by the Prometheus Pod to operate.| `300m`| `300m` | `200m` |
| **prometheusSpec.resources.requests.memory** | Amount of memory requested by the Prometheus Pod to operate. | `1000Mi`| `1Gi` | `200Mi` |
| **alertmanager.alertmanagerSpec.retention** | Time period for which Alertmanager retains data.| `120h` | `240h` | `1h`|

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
    prometheus.prometheusSpec.resources.limits.cpu: "600m"
    prometheus.prometheusSpec.resources.limits.memory: "2Gi"
    prometheus.prometheusSpec.resources.requests.cpu: "300m"
    prometheus.prometheusSpec.resources.requests.memory: "1Gi"
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
      prometheus.prometheusSpec.resources.limits.cpu: "600m"
      prometheus.prometheusSpec.resources.limits.memory: "2Gi"
      prometheus.prometheusSpec.resources.requests.cpu: "300m"
      prometheus.prometheusSpec.resources.requests.memory: "1Gi"
      alertmanager.alertmanagerSpec.retention: "240h"
    EOF
    ```
  2. Run the [cluster update process](/root/kyma/#installation-update-kyma).
  </details>
</div>


