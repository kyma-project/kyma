---
title: Logging production profile
type: Configuration
---
## Overview

To use production-ready Logging, you can install Kyma with the Logging production profile. Higher memory limits set for Loki and FluentBit logging solutions ensure stable log processing in a production environment.

## Parameters

The table shows the parameters used in the production profile and their values:

 Parameter  | Description |  Value   | 
|-----------|-------------|----------|
| **loki.resources.limits.memory** | Maximum amount of memory available for Loki to use. | `512Mi` | 
| **fluent-bit.resources.limits.memory** | Maximum amount of memory available for FluentBit to use. |`256Mi`| 

## Use production profile 

You can deploy a Kyma cluster with Logging configured to use the production profile, or add the configuration in the runtime. Follow these steps:

<div tabs>
  <details>
  <summary>
  Install Kyma with production-ready Logging
 </summary>

1. Create a Kubernetes cluster for Kyma installation.

2. Apply an override that forces Logging to use the production profile:

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
      component: logging
      kyma-project.io/installation: ""
  data:
    loki.resources.limits.memory: "512Mi"
    fluent-bit.resources.limits.memory: "256Mi"
  EOF
  ```
  </details>
  <details>
  <summary>
  Enable configuration in a running cluster
  </summary>

1. Apply an override that forces Logging to use the production profile:

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
      component: logging
      kyma-project.io/installation: ""
  data:
    loki.resources.limits.memory: "512Mi"
    fluent-bit.resources.limits.memory: "256Mi"
  EOF
```
2. Run the [cluster update process](/root/kyma/#installation-update-kyma).
  </details>
</div>


