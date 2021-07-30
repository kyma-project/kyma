---
title: ORY chart
---

To configure the ORY chart, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/ory/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter |  Description | Default value |
|-------|-------|:--------:|
| **oathkeeper.deployment.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **oathkeeper.deployment.resources.requests.cpu** | Defines requests for CPU resources. | `50m` |
| **hpa.oathkeeper.minReplicas** |  Defines the initial number of created Oathkeeper instances. | `1` |
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `true` |
<!--**hydra.hydra.config.secrets.system** | Sets the system encryption string for Hydra. | An at least 16 characters long alphanumerical string |-->
<!--| **hydra.hydra.config.secrets.cookie** | Sets the cookie session encryption string for Hydra. | An at least 16 characters long alphanumerical string |-->
<!--| **global.ory.hydra.persistence.user** | Specifies the name of the user with permissions to access the database. | `dbuser` |-->
<!--| **global.ory.hydra.persistence.secretName** | Specifies the name of the Secret in the same Namespace as Hydra that stores the database password. | `my-secret` |-->
<!--| **global.ory.hydra.persistence.secretKey** | Specifies the name of the key in the Secret that contains the database password. | `my-db-password` |
| **global.ory.hydra.persistence.dbUrl** | Specifies the database URL. For more information, see the [configuration file](https://github.com/ory/hydra/blob/v1.4.1/docs/config.yaml). | `mydb.my-namespace:1234` |
| **global.ory.hydra.persistence.dbName** | Specifies the name of the database saved in Hydra. | `db` |
| **global.ory.hydra.persistence.dbType** | Specifies the type of the database. The supported protocols are `postgres`, `mysql`, `cockroach`. For more information, see the [configuration file](https://github.com/ory/hydra/blob/v1.4.1/docs/config.yaml). | `postgres` |-->
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `true` |
| **global.ory.hydra.persistence.gcloud.enabled** | Defines whether Hydra should initiate the deployment of Google SQL proxy. | `false` |
<!--| **hydra.hydra.config.secrets.system** | Sets the system encryption string for Hydra. | An at least 16 characters long alphanumerical string |
| **hydra.hydra.config.secrets.cookie** | Sets the cookie session encryption string for Hydra. | An at least 16 characters long alphanumerical string |
| **data.global.ory.hydra.persistence.user** | Specifies the name of the user with permissions to access the database. | `dbuser` |
| **data.global.ory.hydra.persistence.secretName** | Specifies the name of the Secret in the same Namespace as Hydra that stores the database password. | `my-secret` |
| **data.global.ory.hydra.persistence.secretKey** | Specifies the name of the key in the Secret that contains the database password. | `my-db-password` |
| **data.global.ory.hydra.persistence.dbUrl** | Specifies the database URL. For more information, see the [configuration file](https://github.com/ory/hydra/blob/v1.4.1/docs/config.yaml). | Required: `ory-gcloud-sqlproxy.kyma-system:DB_PORT` |
| **data.global.ory.hydra.persistence.dbName** | Specifies the name of the database saved in Hydra. | `db` |
| **data.global.ory.hydra.persistence.dbType** | Specifies the type of the database. The supported protocols are `postgres`, `mysql`, `cockroach`. For more information, see the [configuration file](https://github.com/ory/hydra/blob/v1.4.1/docs/config.yaml). | `postgres` |-->
<!-- **gcloud-sqlproxy.cloudsql.instance.instanceName** | Specifies the name of the database instance in GCP. This value is the last part of the string returned by the Cloud SQL Console for **Instance connection name** - the one after the final `:`. For example, if the value for **Instance connection name** is `my_project:my_region:mydbinstance`, use only `mydbinstance`. | `mydbinstance` |
| **gcloud-sqlproxy.cloudsql.instance.project** | Specifies the name of the GCP project used. | `my-gcp-project` |
| **gcloud-sqlproxy.cloudsql.instance.region** | Specifies the name of the GCP **region** used. Note, that it does not equal the GCP **zone**. | `europe-west4` |
| **gcloud-sqlproxy.cloudsql.instance.port** | Specifies the port used by the database to handle connections. Database dependent. | postgres: `5432` mysql: `3306` |-->
| **gcloud-sqlproxy.existingSecret** | Specifies the name of the Secret in the same Namespace as the proxy, that stores the database password. | `ory-hydra-credentials` |
| **gcloud-sqlproxy.existingSecretKey** | Specifies the name of the key in the Secret that contains the [GCP ServiceAccount json key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys). | `gcp-sa.json` |

> **TIP:** See the original [ORY](https://github.com/ory/k8s/tree/master/helm/charts), [ORY Oathkeeper](http://k8s.ory.sh/helm/oathkeeper.html), [PostgreSQL](https://github.com/helm/charts/tree/master/stable/postgresql), and [GCP SQL Proxy](https://github.com/rimusz/charts/tree/master/stable/gcloud-sqlproxy) helm charts for more configuration options.
