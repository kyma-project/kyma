---
title: OAuth2 server profiles
type: Configuration
---

By default, every Kyma deployment is installed with the OAuth2 server using what is considered a default profile. This configuration is not considered production-ready. To use the Kyma OAuth2 server in a production environment, configure Hydra to use the production profile.

## Default profile

In the case of the ORY Hydra OAuth2 server, the default profile includes:
   - An in-cluster database that stores the registered client data.
   - A job that reads the generated database credentials and saves them to the configuration of Hydra before the installation and update.
   - Default resource quotas.


### Persistence mode for the default profile

The default profile for the OAuth2 server enables the use of a [preconfigured PostgreSQL](https://github.com/helm/charts/tree/master/stable/postgresql) database, which is installed together with the Hydra server. The database is created in the cluster as a StatefulSet and uses a PersistentVolume that is provider-specific. This means that the PersistentVolume used by the database uses the default StorageClass of the cluster's host provider. The internal PostgreSQL database is installed with every Kyma deployment and doesn't require manual configuration.

## Production profile

The production profile introduces the following changes to the Hydra OAuth2 server deployment:
   - The registered client data is saved in a user-managed database.
   - Optionally, a Gcloud proxy service is deployed.
   - The Oathkeeper authorization and authentication proxy has raised CPU limits and requests. It starts with more replicas and can scale up horizontally to higher numbers.

### Oathkeeper settings for the production profile

The production profile requires the following parameters in order to operate:

| Parameter |  Description | Required value |
|-------|-------|:--------:|
| **oathkeeper.deployment.resources.limits.cpu** | Defines limits for CPU resources. | `800m` |
| **oathkeeper.deployment.resources.requests.cpu** | Defines requests for CPU resources. | `200m` |
| **hpa.oathkeeper.minReplicas** |  Defines the initial number of created Oathkeeper instances. | `3` |

### Persistence modes for the production profile

The production profile for the OAuth2 server enables the use of a custom database, which can be one of the following options:
  - A user-maintained database to which credentials are supplied.
  - A [GCP Cloud SQL](https://cloud.google.com/sql) instance to which credentials are supplied. In this case, an extra [gcloud-proxy](https://cloud.google.com/sql/docs/mysql/sql-proxy) deployment is created allowing secured connections.

#### Custom database

Alternatively, you can use a compatible, custom database to store the registered client data. To use a database, you must create a Kubernetes Secret with the database password as an override for the Hydra OAuth2 server. The details of the database are passed using these parameters of the production profile override:

**General settings:**

| Parameter | Description | Required value |
|----------|------| :---: |
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `false` |
| **hydra.hydra.config.secrets.system** | Sets the system encryption string for Hydra. | An at least 16 characters long alphanumerical string |
| **hydra.hydra.config.secrets.cookie** | Sets the cookie session encryption string for Hydra. | An at least 16 characters long alphanumerical string |

**Database settings:**

| Parameter | Description | Example value |
|----------|------| :---: |
| **global.ory.hydra.persistence.user** | Specifies the name of the user with permissions to access the database. | `dbuser` |
| **global.ory.hydra.persistence.secretName** | Specifies the name of the Secret in the same Namespace as Hydra that stores the database password. | `my-secret` |
| **global.ory.hydra.persistence.secretKey** | Specifies the name of the key in the Secret that contains the database password. | `my-db-password` |
| **global.ory.hydra.persistence.dbUrl** | Specifies the database URL. For more information, see the [configuration file](https://github.com/ory/hydra/blob/v1.4.1/docs/config.yaml). | `mydb.my-namespace:1234` |
| **global.ory.hydra.persistence.dbName** | Specifies the name of the database saved in Hydra. | `db` |
| **global.ory.hydra.persistence.dbType** | Specifies the type of the database. The supported protocols are `postgres`, `mysql`, `cockroach`. For more information, see the [configuration file](https://github.com/ory/hydra/blob/v1.4.1/docs/config.yaml). | `postgres` |

#### Google Cloud SQL

The Cloud SQL is a provider-supplied and maintained database, which requires a special proxy deployment in order to provide a secured connection. In Kyma we provide a [pre-installed](https://github.com/rimusz/charts/tree/master/stable/gcloud-sqlproxy) deployment, which requires the following parameters in order to operate:

**General settings:**

| Parameter | Description | Required value |
|----------|------| :---: |
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `false` |
| **global.ory.hydra.persistence.gcloud.enabled** | Defines whether Hydra should initiate the deployment of Google SQL proxy. | `true` |
| **hydra.hydra.config.secrets.system** | Sets the system encryption string for Hydra. | An at least 16 characters long alphanumerical string |
| **hydra.hydra.config.secrets.cookie** | Sets the cookie session encryption string for Hydra. | An at least 16 characters long alphanumerical string |

**Database settings:**

| Parameter | Description | Example value |
|----------|------| :---: |
| **data.global.ory.hydra.persistence.user** | Specifies the name of the user with permissions to access the database. | `dbuser` |
| **data.global.ory.hydra.persistence.secretName** | Specifies the name of the Secret in the same Namespace as Hydra that stores the database password. | `my-secret` |
| **data.global.ory.hydra.persistence.secretKey** | Specifies the name of the key in the Secret that contains the database password. | `my-db-password` |
| **data.global.ory.hydra.persistence.dbUrl** | Specifies the database URL. For more information, see the [configuration file](https://github.com/ory/hydra/blob/v1.4.1/docs/config.yaml). | Required: `ory-gcloud-sqlproxy.kyma-system:DB_PORT` |
| **data.global.ory.hydra.persistence.dbName** | Specifies the name of the database saved in Hydra. | `db` |
| **data.global.ory.hydra.persistence.dbType** | Specifies the type of the database. The supported protocols are `postgres`, `mysql`, `cockroach`. For more information, see the [configuration file](https://github.com/ory/hydra/blob/v1.4.1/docs/config.yaml). | `postgres` |

**Proxy settings:**

| Parameter | Description | Example value |
|----------|------| :---: |
| **gcloud-sqlproxy.cloudsql.instance.instanceName** | Specifies the name of the database instance in GCP. This value is the last part of the string returned by the Cloud SQL Console for **Instance connection name** - the one after the final `:`. For example, if the value for **Instance connection name** is `my_project:my_region:mydbinstance`, use only `mydbinstance`. | `mydbinstance` |
| **gcloud-sqlproxy.cloudsql.instance.project** | Specifies the name of the GCP project used. | `my-gcp-project` |
| **gcloud-sqlproxy.cloudsql.instance.region** | Specifies the name of the GCP **region** used. Note, that it does not equal the GCP **zone**. | `europe-west4` |
| **gcloud-sqlproxy.cloudsql.instance.port** | Specifies the port used by the database to handle connections. Database dependent. | postgres: `5432` mysql: `3306` |
| **gcloud-sqlproxy.existingSecret** | Specifies the name of the Secret in the same Namespace as the proxy, that stores the database password. | `my-secret` |
| **gcloud-sqlproxy.existingSecretKey** | Specifies the name of the key in the Secret that contains the [GCP ServiceAccount json key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys). | `sa.json` |

>**NOTE:** When using any kind of custom database (gcloud, or self-maintained), it is important to provide the **hydra.hydra.config.secrets** variables, otherwise a random secret will be generated. This secret needs to be common to all Hydra instances using the same instance of the chosen database.

## Use the production profile

Follow these steps to migrate your Oauth2 server to the production profile:

1. Apply an override that forces the Hydra OAuth2 server to use the database of your choice. Follow these links to find an example of override data for each persistence mode:
- [User-maintained](assets/003-ory-db-custom.yaml)
- [Google Cloud SQL](assets/004-ory-db-gcloud.yaml)

2. Run the [cluster update process](/root/kyma/#installation-update-kyma).

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

>**TIP:** All the client data registered by Hydra Maester is migrated to the new database as a part of the update process. During this process, the clients will not be available which may result in errors on issuing the token. If you notice missing or inconsistent data, delete the Hydra Maester Pod to force reconciliation. For more information, read about [Hydra Maester controller and Oauth2 client registration in Kyma](#details-o-auth2-and-open-id-connect-server).
