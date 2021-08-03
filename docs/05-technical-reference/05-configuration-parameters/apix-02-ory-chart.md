---
title: ORY chart
---

To configure the ORY chart and its sub-charts, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/ory/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter |  Description | Default value |
|-------|-------|:--------:|
| **oathkeeper.deployment.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **oathkeeper.deployment.resources.requests.cpu** | Defines requests for CPU resources. | `50m` |
| **hpa.oathkeeper.minReplicas** |  Defines the initial number of created Oathkeeper instances. | `1` |
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `true` |
| **gcloud-sqlproxy.existingSecret** | Specifies the name of the Secret in the same Namespace as the proxy, that stores the database password. | `ory-hydra-credentials` |
| **gcloud-sqlproxy.existingSecretKey** | Specifies the name of the key in the Secret that contains the [GCP ServiceAccount json key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys). | `gcp-sa.json` |

> **TIP:** See the original [ORY](https://github.com/ory/k8s/tree/master/helm/charts), [ORY Oathkeeper](http://k8s.ory.sh/helm/oathkeeper.html), [PostgreSQL](https://github.com/helm/charts/tree/master/stable/postgresql), and [GCP SQL Proxy](https://github.com/rimusz/charts/tree/master/stable/gcloud-sqlproxy) helm charts for more configuration options.
