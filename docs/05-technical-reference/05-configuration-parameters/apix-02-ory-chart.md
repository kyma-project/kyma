---
title: ORY chart
---

To configure the ORY chart and its sub-charts, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/ory/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter |  Description | Default value |
|-------|-------|:--------:|
| **global.postgresql.postgresqlDatabase** | Specifies the name of the database saved in Hydra. | `db4hydra` |
| **global.postgresql.postgresqlUsername** | Specifies the username of the database saved in Hydra. | `hydra` |
| **global.istio.gateway.name** | Specifies the name of the Istio Gateway used in Hydra. | `kyma-gateway` |
| **global.istio.gateway.namespace** | Specifies the Namespace of the Istio Gateway used in Hydra. | `kyma-system` |
| **global.ory.oathkeeper.maester.mode** | Specifies the mode in which ORY Oathkeeper Maester is expected to be deployed. | `sidecar` |
| **global.ory.hydra.persistence.enabled** | Sets persistence for Hydra. | `true`|
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `true` |
| **hpa.oathkeeper.minReplicas** |  Defines the initial number of created Oathkeeper instances. | `1` |
| **hpa.oathkeeper.maxReplicas** |  Defines the maximum number of created Oathkeeper instances. | `3` |
| **hydra.deployment.resources.limits.cpu** | Defines limits for CPU resources. | `500m` |
| **hydra.deployment.resources.limits.memory** | Defines limits for memory resources. | `256Mi` |
| **hydra.deployment.resources.requests.cpu** | Defines requests for CPU resources. | `100m` |
| **hydra.deployment.resources.requests.memory** | Defines requests for memory resources. | `128Mi` |
| **hydra.hydra.existingSecret** | Specifies the name of an existing Kubernetes Secret containing credentials required for Hydra. A default Secret with random values is generated if this value is not provided. | `"ory-hydra-credentials"` |
| **oathkeeper.deployment.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **oathkeeper.deployment.resources.limits.memory** | Defines limits for memory resources.| `128Mi` |
| **oathkeeper.deployment.resources.requests.cpu** | Defines requests for CPU resources. | `50m` |
| **oathkeeper.deployment.resources.requests.memory** | Defines requests for memory resources. | `64Mi` |
| **oathkeeper.oathkeeper-maester.deployment.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **oathkeeper.oathkeeper-maester.deployment.resources.limits.memory** | Defines limits for memory resources. | `50Mi` |
| **oathkeeper.oathkeeper-maester.deployment.resources.requests.cpu** | Defines requests for CPU resources. | `50m` |
| **oathkeeper.oathkeeper-maester.deployment.resources.requests.memory** | Defines requests for memory resources. | `20Mi` |
| **postgresql.resources.requests.memory** | Defines requests for memory resources. | `256Mi` |
| **postgresql.resources.requests.cpu** | Defines requests for CPU resources. | `250m` |
| **postgresql.resources.limits.memory** | Defines limits for memory resources.| `1024Mi` |
| **postgresql.resources.limits.cpu** | Defines limits for CPU resources. | `750m` |
| **postgresql.existingSecret** | Specifies the name of an existing secret to use for PostgreSQL passwords. | `"ory-hydra-credentials"` |
| **gcloud-sqlproxy.existingSecret** | Specifies the name of the Secret in the same Namespace as the proxy, that stores the database password. | `ory-hydra-credentials` |
| **gcloud-sqlproxy.existingSecretKey** | Specifies the name of the key in the Secret that contains the [GCP ServiceAccount json key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys). | `gcp-sa.json` |

> **TIP:** See the original [ORY](https://github.com/ory/k8s/tree/master/helm/charts), [ORY Oathkeeper](http://k8s.ory.sh/helm/oathkeeper.html), [PostgreSQL](https://github.com/helm/charts/tree/master/stable/postgresql), and [GCP SQL Proxy](https://github.com/rimusz/charts/tree/master/stable/gcloud-sqlproxy) helm charts for more configuration options.
