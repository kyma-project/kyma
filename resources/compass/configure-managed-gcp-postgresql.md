# Configure Managed GCP PostgreSQL

## Prerequisites

To set up the database, create the following:

* Cloud SQL Postgres instance on Google Cloud Platform with public IP assigned
* Key for GCP Service Account with at least one of the following roles ([click here for more information](https://cloud.google.com/sql/docs/postgres/connect-external-app#4_if_required_by_your_authentication_method_create_a_service_account)):
  - Cloud SQL Client
  - Cloud SQL Editor
  - Cloud SQL Admin

## Install Compass with managed GCP PostgreSQL database

To install Compass with GCP managed Postgres database, set the **database.embedded.enabled** value to `false` inside the `./chart/compass/values.yaml` file, and fill these values:

| Parameter | Description | Values | Default |
| --- | --- | --- | --- |
| `global.database.managedGCP.serviceAccountKey` | Specifies base64 encoded the key for GCP Service Account mentioned in prerequisites. | base64 encoded string | "" |
| `global.database.managedGCP.instanceConnectionName` | Specifies instance connection name to GCP PostgreSQL database | string | "" |
| `global.database.managedGCP.dbUser` | Specifies database username | string | "" |
| `global.database.managedGCP.dbPassword` | Specifies password for database user | string | "" |
| `global.database.managedGCP.directorDBName` | Specifies Director database name | string | "" |
| `global.database.managedGCP.provisionerDBName` | Specifies Provisioner database name | string | "" |
| `global.database.managedGCP.host` | Specifies cloudsql-proxy host | string | "localhost" |
| `global.database.managedGCP.hostPort` | Specifies cloudsql-proxy port | string | "5432" |
| `global.database.managedGCP.sslMode` | Specifies SSL connection mode | string | "" |

To connect to managed database, we use [cloudsql-proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy) provided by Google, which consumes `serviceAccountKey` and `instanceConnectionName` values.

To find `Instance connection name`, go to the [SQL Instances page](https://console.cloud.google.com/sql/instances) and open desired database overview.
![Instance connection String](./assets/sql-instances-list.png)

Than look for `Instance connection name` box inside `Connect to this instance` section.

![Instance connection String](./assets/instance-connection-string.png)

For `dbUser` and `dbPassword` values, use one of accounts from `USERS` tab. The `dbName` value is name of database which you want to use. The available names can be found inside `DATABASES` tab.

The `host` and the `hostPort` values specifies the cloudsql-proxy host and port. These are used directly by application to connect to proxy, and further to database.

The `sslMode` value specifies SSL connection mode to database. Check possible values [here](https://www.postgresql.org/docs/9.1/libpq-ssl.html) under `SSL Mode Descriptions` section.
