---
title: ORY chart
---

To configure the ORY chart and its sub-charts, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/ory/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter |  Description | Default value |
|-------|-------|:--------:|
| **global.ory.oathkeeper.maester.mode** | Specifies the mode in which ORY Oathkeeper Maester is expected to be deployed. | `sidecar` |
| **hpa.oathkeeper.minReplicas** |  Defines the initial number of created Oathkeeper instances. | `1` |
| **hpa.oathkeeper.maxReplicas** |  Defines the maximum number of created Oathkeeper instances. | `3` |
| **oathkeeper.deployment.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **oathkeeper.deployment.resources.limits.memory** | Defines limits for memory resources.| `128Mi` |
| **oathkeeper.deployment.resources.requests.cpu** | Defines requests for CPU resources. | `50m` |
| **oathkeeper.deployment.resources.requests.memory** | Defines requests for memory resources. | `64Mi` |
| **oathkeeper.oathkeeper-maester.deployment.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **oathkeeper.oathkeeper-maester.deployment.resources.limits.memory** | Defines limits for memory resources. | `50Mi` |
| **oathkeeper.oathkeeper-maester.deployment.resources.requests.cpu** | Defines requests for CPU resources. | `50m` |
| **oathkeeper.oathkeeper-maester.deployment.resources.requests.memory** | Defines requests for memory resources. | `20Mi` |
| **gcloud-sqlproxy.existingSecretKey** | Specifies the name of the key in the Secret that contains the [GCP service account JSON key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys). | `gcp-sa.json` |

> **TIP:** See the original [ORY](https://github.com/ory/k8s/tree/master/helm/charts) and [ORY Oathkeeper](http://k8s.ory.sh/helm/oathkeeper.html) helm charts for more configuration options.
