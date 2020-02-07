---
title: OAuth2 server production profile
type: Configuration
---

By default, every Kyma deployment is installed with the OAuth2 server using what is considered a development profile. In the case of the ORY Hydra OAuth2 server, this means that:
  - Hydra works in the "in-memory" mode and saves the registered OAuth2 clients in its memory, without the use of persistent storage.
  - Default resource quotas are used.

This configuration is not considered production-ready. To use the Kyma OAuth2 server in a production environment, configure Hydra to use the production profile.

## The production profile

The production profile introduces the following changes to the Hydra OAuth2 server deployment:
   - Persistence is enabled for Hydra; the registered client data is saved in a in-cluster database or in a user-managed database.
   - A job that reads the generated database credentials and saves them to the configuration of Hydra is added.
   - Optionally, a Gcloud proxy service is deployed.

### Persistence modes

The production profile for the OAuth2 server enables the use of a database, which can be one of the following options:
  - A [preconfigured PostgreSQL](https://github.com/helm/charts/tree/master/stable/postgresql) database, which is installed together with the Hydra server.
  - A user-maintained database to which credentials are supplied.
  - A [GCP Cloud SQL](https://cloud.google.com/sql) instance to which credentials are supplied. In this case, an extra [gcloud-proxy](https://cloud.google.com/sql/docs/mysql/sql-proxy) deployment is created allowing secured connections.

#### Internal PostgreSQL setup
The database is created in the cluster as a StatefulSet and uses a PersistentVolume that is provider-specific. This means that the PersistentVolume used by the database uses the default StorageClass of the cluster's host provider.

In order to use the database, the following configuration parameters have to be supplied:

**General settings:**

| Parameter | Description | Required value |
|----------|-----------| :---: |
| **global.ory.hydra.persistence.enabled** | Defines whether Hydra should use the `database` mode of operation. If `false`, it uses the `in-memory` mode. | `true` |
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `true` |
| **hydra.hydra.autoMigrate** | Enables Hydra auto-migration feature, which prepares the database | `true` | 

You can find an example of the required [configmap here](./assets/003-ory-db-postgres.yaml)

#### Custom database

Alternatively, you can use a compatible, custom database to store the registered client data. To use a database, you must create a Kubernetes Secret with the database password as an override for the Hydra OAuth2 server. The details of the database are passed using these parameters of the production profile override:

**General settings:**

| Parameter | Description | Required value |
|----------|------| :---: |
| **global.ory.hydra.persistence.enabled** | Defines whether Hydra should use the `database` mode of operation. If false, it uses the `in-memory` mode. | `true` |
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `false` |
| **hydra.hydra.autoMigrate** | Enables Hydra auto-migration feature, which prepares the database. | `true` | 
| **hydra.hydra.config.secrets.system** | Sets the system encryption string for Hydra. | An at least 16 characters long alphanumerical string | 
| **hydra.hydra.config.secrets.cookie** | Sets the cookie session encryption string for Hydra. | An at least 16 characters long alphanumerical string |

**Database settings:**

| Parameter | Description | Example value |
|----------|------| :---: |
| **global.ory.hydra.persistence.user** | Specifies the name of the user with permissions to access the database. | `dbuser` |
| **global.ory.hydra.persistence.secretName** | Specifies the name of the Secret in the same Namespace as Hydra that stores the database password. | `my-secret` |
| **global.ory.hydra.persistence.secretKey** | Specifies the name of the key in the Secret that contains the database password. | `my-db-password` |
| **global.ory.hydra.persistence.dbUrl** | Specifies the database URL. For more information, read [this](https://github.com/ory/hydra/blob/master/docs/config.yaml) document. | `mydb.mynamespace.svc.cluster.local:1234` |
| **global.ory.hydra.persistence.dbName** | Specifies the name of the database saved in Hydra. | `db` |
| **global.ory.hydra.persistence.dbType** | Specifies the type of the database. The supported protocols are `postgres`, `mysql`, `cockroach`. Follow [this](https://github.com/ory/hydra/blob/master/docs/config.yaml) link for more information. | `postgres` |

You can find an example of the required configmap [here](./assets/004-ory-db-custom.yaml)

#### Google Cloud SQL

The Cloud SQL is a provider-supplied and maintained database, which requires a special proxy deployment in order to provide a secured connection. In Kyma we provide a [pre-installed](https://github.com/rimusz/charts/tree/master/stable/gcloud-sqlproxy) deployment, which requires the following parameters in order to operate:

**General settings:**

| Parameter | Description | Required value |
|----------|------| :---: |
| **global.ory.hydra.persistence.enabled** | Defines whether Hydra should use the `database` mode of operation. If set to `false`, it uses the `in-memory` mode. | `true` |
| **global.ory.hydra.persistence.postgresql.enabled** | Defines whether Hydra should initiate the deployment of an in-cluster database. Set to `false` to use a self-provided database. If set to `true`, Hydra always uses an in-cluster database and ignores the custom database details. | `false` |
| **hydra.hydra.autoMigrate** | Enables Hydra auto-migration feature, which prepares the database. | `true` | 
| **hydra.hydra.config.secrets.system** | Sets the system encryption string for Hydra. | An at least 16 characters long alphanumerical string | 
| **hydra.hydra.config.secrets.cookie** | Sets the cookie session encryption string for Hydra. | An at least 16 characters long alphanumerical string |

**Database settings:**

| Parameter | Description | Example value |
|----------|------| :---: |
| **data.global.ory.hydra.persistence.user** | Specifies the name of the user with permissions to access the database. | `dbuser` |
| **data.global.ory.hydra.persistence.secretName** | Specifies the name of the Secret in the same Namespace as Hydra that stores the database password. | `my-secret` |
| **data.global.ory.hydra.persistence.secretKey** | Specifies the name of the key in the Secret that contains the database password. | `my-db-password` |
| **data.global.ory.hydra.persistence.dbUrl** | Specifies the database URL. For more information, read [this](https://github.com/ory/hydra/blob/master/docs/config.yaml) document. | Required: `ory-gcloud-sqlproxy.kyma-system.svc.cluster.local:5432` |
| **data.global.ory.hydra.persistence.dbName** | Specifies the name of the database saved in Hydra. | `db` |
| **data.global.ory.hydra.persistence.dbType** | Specifies the type of the database. The supported protocols are `postgres`, `mysql`, `cockroach`. Follow [this](https://github.com/ory/hydra/blob/master/docs/config.yaml) link for more information. | `postgres` |

**Proxy settings:**

| Parameter | Description | Example value |
|----------|------| :---: |
| **gcloud-sqlproxy.cloudsql.instance.instanceName** | Specifies the name of the database instance in GCP. | `mydbinstance` |
| **gcloud-sqlproxy.cloudsql.instance.project** | Specifies the name of the GCP project used. | `my-gcp-project` |
| **gcloud-sqlproxy.cloudsql.instance.region** | Specifies the name of the GCP **region** used. Note, that it does not equal the GCP **zone**. | `europe-west4` |
| **gcloud-sqlproxy.cloudsql.instance.port** | Specifies the port used by the database to handle connections. Database dependent. | postgres: `5432` mysql: `3306` |
| **gcloud-sqlproxy.existingSecret** | Specifies the name of the Secret in the same Namespace as the proxy, that stores the database password. | `my-secret` |
| **gcloud-sqlproxy.existingSecretKey** | Specifies the name of the key in the Secret that contains the [GCP ServiceAccount json key](https://cloud.google.com/iam/docs/creating-managing-service-account-keys). | `sa.json` |

You can find an example of the required configmap [here](./assets/005-ory-db-gcloud.yaml).

>>**NOTE:** When using any kind of custom database (gcloud, or self-maintained), it is important to set the 

## Use the production profile

You can deploy a Kyma cluster with the Hydra OAuth2 server configured to use the production profile, or configure Hydra in a running cluster to use the production profile. Follow these steps:

<div tabs>
  <details>
  <summary>
  Install Kyma with production-ready Hydra
  </summary>
  >**NOTE:** This configuration installs a PorstgreSQL database in the Kyma cluster.

  1. Create an appropriate Kubernetes cluster for Kyma in your host environment.
  2. Apply an override that forces the Hydra OAuth2 server to use the production profile. Run:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: ory-overrides
      namespace: kyma-installer
      labels:
        installer: overrides
        component: ory
        kyma-project.io/installation: ""
    data:
      global.ory.hydra.persistence.enabled: "true"
      global.ory.hydra.persistence.postgresql.enabled: "true"
      global.ory.hydra.persistence.gcloud.enabled: "false"
      hydra.hydra.autoMigrate: "true"
    EOF
    ```
  3. Install Kyma on the cluster.

  </details>
  <details>
  <summary>
  Enable production profile in a running cluster
  </summary>

  >**CAUTION:** When you configure Hydra to use the production profile in a running cluster, you lose all registered clients. Using the production profile restarts the Hydra Pod, which wipes the entire "in-memory" storage used to save the registered client data by default.

  >**NOTE:** This configuration installs a PorstgreSQL database in the Kyma cluster.

  1. Apply an override that forces the Hydra OAuth2 server to use the production profile. Run:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: ory-overrides
      namespace: kyma-installer
      labels:
        installer: overrides
        component: ory
        kyma-project.io/installation: ""
    data:
      global.ory.hydra.persistence.enabled: "true"
      global.ory.hydra.persistence.postgresql.enabled: "true"
      global.ory.hydra.persistence.gcloud.enabled: "false"
      hydra.hydra.autoMigrate: "true"
    EOF
    ```
  2. Run the cluster [update procedure](/root/kyma/#installation-update-kyma).


  </details>

</div>
