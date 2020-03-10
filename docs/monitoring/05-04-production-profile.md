---
title: Monitoring profiles
type: Configuration
---

When you install Kyma, Monitoring uses settings defined in what is considered a development profile. In such a configuration, Prometheus stores data only for one day which may cause a problem when more time is needed to identify and solve an issue.
To make Monitoring settings production-ready and avoid potential performance issues, configure Monitoring to use the production profile.  

## Production profile 

The production profile introduces the following changes to monitoring: 

* Increased retention time to prevent data loss in case of prolonged troubleshooting. 
* Increased memory and CPU values to ensure stable performance. 

When you deploy a Kyma cluster with a production profile, the override includes the following parameters:

 Parameter  | Description  |  Value        |
|-----------|-------------|---------------|
| **retentionSize** | Maximum number of bytes that storage blocks can use. The oldest data will be removed first. | `15GB` |
| **retention** | Period for which Prometheus stores metrics in-memory. This retention time applies to in-memory storage only. Prometheus stores the recent data in-memory for the specified amount of time to avoid reading the entire data from disk.| `30d` |
| **prometheusSpec.volumeClaimTemplate.spec.resources.requests.storage** | Amount of storage requested by the Prometheus Pod. | `20Gi` |
| **prometheusSpec.resources.limits.cpu** | Maximum number of CPUs that will be made available for Prometheus Pod to use | `600m` |
| **prometheusSpec.resources.limits.memory** | Maximum amount of memory that will be made available for the Prometheus Pod to use. | `2Gi` |
| **prometheusSpec.resources.requests.cpu** |  Number of CPUs requested by the Prometheus Pod to operate.| `300m` |
| **prometheusSpec.resources.requests.memory** | Amount of memory requested by the Prometheus Pod to operate. | `1Gi` |
| **prometheusSpec.alertmanagerSpec.retention** | Time duration Alertmanager retains data for.  | `240h` |

### Use the production profile

You can deploy a Kyma cluster with Monitoring configured to use the production profile, or add the configuration in the runtime. Follow these steps:

<div tabs>
  <details>
  <summary>
  Install Kyma with production-ready Monitoring
 </summary>

Follow these steps to update the settings for Prometheus and Alertmanager.

1. Create an appropriate Kubernetes cluster for Kyma in your host environment.

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
  2. Run the [cluster update procedure](/root/kyma/#installation-update-kyma).

  </details>
</div>

## Local profile 

If you deploy Kyma locally on Minikube, Monitoring uses lightweight [local configuration](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-config-local.yaml.tpl). This configuration provides lower retention time, memory, and CPU values to avoid high memory consumption and ensure stable performance. 

