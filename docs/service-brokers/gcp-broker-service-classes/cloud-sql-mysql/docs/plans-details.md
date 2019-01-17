---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Beta Plan` | Cloud SQL-MySQL plan for the Beta release of the Google Cloud Platform Service Broker |

## Provisioning parameters

Provisioning an instance creates a new MySQL instance. These are the input parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **databaseVersion** | `string` | The database engine type and version. The value can be either `MYSQL_5_7` or `MYSQL_5_6`. The choice is permanent. | NO | `MYSQL_5_7` |
| **failoverReplica** | `object` | The name and status of the failover replica. This property is applicable only to Second Generation instances. | NO | - |
| **instanceId** | `string` | CloudSQL instance ID. Use lowercase letters, numbers, and hyphens. Start with a letter. Must be 1-78 characters long. The choice is permanent. | YES | - |
| **masterInstanceName** | `string` | The name of the instance which acts as master in the replication setup. | NO | - |
| **onPremisesConfiguration** | `object` | Configuration specific to on-premises instances.| NO | - |
| **onPremisesConfiguration.hostPort** | `string` | The host and port of the on-premises instance in the `host:port` format | NO | - |
| **region** | `string` | Determines where your CloudSQL data is located. For better performance, keep your data close to the services that need it. These are the possible values: `asia-east1`, `asia-northeast1`, `asia-south1`, `australia-southeast1`, `europe-west1`, `europe-west2`, `europe-west3`, `europe-west4`, `northamerica-northeast1`, `southamerica-east1`, `us-central1`, `us-east1`, `us-east4`, `us-west1`. The choice is permanent.| NO | `us-central1` |
| **replicaConfiguration** | `object` | Configuration specific to read-replicas replicating from on-premises masters. For more information, go to the **ReplicaConfiguration properties** section.  | NO | - |
| **settings** | `object` | The user settings. For more information, go to the **Settings properties** section. | YES | - |

### ReplicaConfiguration properties

These are the **replicaConfiguration** properties:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **failoverTarget** | `boolean` | Specifies if the replica is the failover target. If the field is set to `true`, the replica is designated as a failover replica. In case the master instance fails, the replica instance is promoted as the new master instance. Only one replica can be specified as a failover target and this replica must be in a different zone with the master instance. | NO | - |
| **mysqlReplicaConfiguration** | `object` | MySQL specific configuration when replicating from a MySQL on-premises master. Replication configuration information, such as the username, password, certificates, and keys, are not stored in the instance metadata. The configuration information is used only to set up the replication connection and is stored by MySQL in the `master.info` file in the data directory. For more information, go to the **MySqlReplicaConfiguration properties** section. | NO | - |

### MySqlReplicaConfiguration properties

These are the **MySqlReplicaConfiguration** properties:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **caCertificate** | `string` | PEM representation of the trusted CA's x509 certificate. | NO | - |
| **clientCertificate** | `string` | PEM representation of the slave's x509 certificate. | NO | - |
| **clientKey** | `string` | PEM representation of the slave's private key. The corresponding public key is encoded in the client's certificate. | NO | - |
| **connectRetryInterval** | `integer` | Seconds to wait between connect retries. | NO | `60 seconds` |
| **dumpFilePath** | `string` | Path to an SQL dump file in Google Cloud Storage from which the slave instance is created. The URI is in the `gs://{bucketName}/{fileName}` form. Compressed gzip files (.gz) are also supported. Dumps should have the binlog co-ordinates from which replication should begin. This can be accomplished by setting **--master-data** to `1` when using mysqldump. | NO | - |
| **masterHeartbeatPeriod** | `string` | Interval in milliseconds between replication heartbeats. | NO | - |
| **password** | `string` | The password for the replication connection. | NO | - |
| **sslCipher** | `string` | A list of permissible ciphers to use for SSL encryption. | NO | - |
| **username** | `string` | The username for the replication connection. | NO | - |
| **verifyServerCertificate** | `boolean` | Whether or not to check the master's Common Name value in the certificate that it sends during the SSL handshake. | NO | - |

### Settings properties

These are the **Settings** properties:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **activationPolicy** | `string` | The activation policy specifies when the instance is activated. It is applicable only when the instance state is `RUNNABLE`. The possible values are `ALWAYS`, `NEVER` and `ON_DEMAND`. `ALWAYS` indicates that the instance is on, and remains so even in the absence of connection requests. `NEVER` means that the instance is off and is not activated, even if a connection request arrives. `ON_DEMAND` applies to First Generation instances only and it indicates that the instance responds to incoming requests, and turns itself off when not in use. Instances with `PER_USE` pricing turn off after 15 minutes of inactivity. Instances with `PER_PACKAGE` pricing turn off after 12 hours of inactivity. | NO | - |
| **authorizedGaeApplications** | `string` | The App Engine application IDs that can access this instance. This property is only applicable to First Generation instances. | NO | - |
| **backupConfiguration** | `string` | The daily backup configuration for the instance. For more information, go to the **BackupConfiguration properties** section. | NO | - |
| **crashSafeReplicationEnabled** | `boolean` | Configuration specific to read replica instances. Indicates whether database flags for crash-safe replication are enabled. This property is only applicable to First Generation instances. | NO | - |
| **dataDiskSizeGb** | `string` | The size of data disk in `GB`. The data disk size minimum is `10GB`. Applies only to Second Generation instances. | NO | - |
| **dataDiskType** | `string` | The type of data disk. The possible values are `PD_SSD`, `PD_HDD`. Applies only to Second Generation instances. | NO | `PD_SSD` |
| **databaseFlags** | `array` | The database flags passed to the instance at startup. | NO | - |
| **databaseFlags.name** | `string` | The name of the flag. These flags are passed at instance startup, so include both MySQL server options and MySQL system variables. Flags should be specified with underscores, not hyphens. | NO | - |
| **databaseFlags.value** | `string` | The value of the flag. Booleans should be set to `on` for `true` and `off` for `false`. This field must be omitted if the flag does not take a value. | NO | - |
| **databaseReplicationEnabled** | `boolean` | Configuration specific to read replica instances. Indicates whether replication is enabled or not. | NO | - |
| **ipConfiguration** | `object` | The settings for IP Management. This allows to enable or disable the instance IP and manage which external networks can connect to the instance. The IPv4 address cannot be disabled for Second Generation instances. For more information, go to the **IpConfiguration properties** section. | NO | - |
| **locationPreference** | `object` | The location preference settings. This allows the instance to be located as near as possible to either an App Engine application or Compute Engine zone for better performance. App Engine co-location is only applicable to First Generation instances. | NO | - |
| **locationPreference.followGaeApplication** | `string` | The AppEngine application to follow. It must be in the same region as the Cloud SQL instance. | NO | - |
| **locationPreference.zone** | `string` | The preferred Compute Engine zone. | NO | - |
| **maintenanceWindow** | `object` | The maintenance window for this instance. This specifies when the instance may be restarted for maintenance purposes. Applies only to Second Generation instances. | NO | - |
| **maintenanceWindow.day** | `integer` | Day of week (1-7), starting on Monday. | NO | - |
| **maintenanceWindow.hour** | `integer` | The hour of the day  (0-23). | NO | - |
| **maintenanceWindow.updateTrack** | `string` | Maintenance timing setting: canary or stable. | NO | - |
| **pricingPlan** | `string` | The pricing plan for this instance. The value can be either `PER_USE` or `PACKAGE`. Only `PER_USE` is supported for Second Generation instances. | NO | `PER_USE` |
| **replicationType** | `string` | The type of replication this instance uses. This can be either `ASYNCHRONOUS` or `SYNCHRONOUS`. This property is only applicable to First Generation instances. | NO | - |
| **settingsVersion** | `string` | The version of instance settings. This is a required field for update method to make sure concurrent updates are handled properly. During update, use the most recent **settingsVersion** value for this instance and do not try to update this value. | NO | - |
| **storageAutoResize** | `boolean` | Configuration to increase storage size automatically. Applies only to Second Generation instances. | NO | `true` |
| **storageAutoResizeLimit** | `string` | The maximum size to which storage capacity can be automatically increased. The default value is `0`, which specifies that there is no limit. Applies only to Second Generation instances. | NO | `0` |
| **tier** | `string` | For better performance, choose a CloudSQL machine type with enough memory to hold your largest table. These are the possible values: `db-f1-micro`, `db-g1-small`, `db-n1-standard-1`, `db-n1-standard-2`, `db-n1-standard-4`, `db-n1-standard-8`, `db-n1-standard-16`, `db-n1-standard-32`, `db-n1-standard-64`, `db-n1-highmem-2`, `db-n1-highmem-4`, `db-n1-highmem-8`, `db-n1-highmem-16`, `db-n1-highmem-32`, `db-n1-highmem-64` | YES | `db-n1-standard-1` |
| **userLabels** | `object` | To organize your project, add arbitrary labels as key/value pairs to CloudSQL. Use labels to indicate different elements, such as Namespaces, services, teams. | NO | - |

### BackupConfiguration properties

These are the **BackupConfiguration** properties:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **enabled** | `boolean` | Indicates if this configuration is enabled. | NO | - |
| **binaryLogEnabled** | `boolean` | Indicates if the binary log is enabled. If backup configuration is disabled, binary log must be disabled as well. | NO | - |
| **startTime** | `string` | Start time for the daily backup configuration in UTC timezone in the 24 hour, `HH:MM`, format. | NO | - |

### IpConfiguration properties

These are the **IpConfiguration** properties:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **authorizedNetworks** | `array` | The list of external networks that are allowed to connect to the instance using the IP. In CIDR notation, also known as slash notation. | NO | - |
| **authorizedNetworks.expirationTime** | `string` | The time in the RFC 3339 format when this access control entry expires. | NO | - |
| **authorizedNetworks.name** | `string` | An optional label to identify this entry. | NO | - |
| **authorizedNetworks.value** | `string` | The whitelisted value for the access control list. | NO | - |
| **ipv4Enabled** | `boolean` | Indicates whether the instance should be assigned an IP address or not. | NO | - |
| **requireSsl** | `boolean` | Indicates whether SSL connections over IP should be enforced or not. | NO | - |

## Update parameters

The update parameters are the same as the provisioning parameters.

## Binding parameters

Binding grants one of the available IAM roles on the Cloud SQL instance to the specified service account. Optionally, you can create a new service account and add the access to the MySQL instance. These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **createServiceAccount** | `boolean` | Create a new service account for MySQL binding. | NO | `false` |
| **roles** | `array` | The list of CloudSQL roles for the binding. Affects the level of access granted to the service account. The value of this parameter is `roles/cloudsql.client`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | `roles/cloudsql.client` |
| **serviceAccount** | `string` | The GCP service account to which access is granted. | YES | - |

### Credentials

Binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **connectionName** | `string` | The SQL instance name. |
| **privateKeyData** | `JSON Object` | The service account OAuth information. |
| **serviceAccount** | `string` | The GCP service account to which access is granted. |

### Usage

To connect your business application to your Cloud SQL instance, use the [Cloud SQL proxy](https://github.com/GoogleCloudPlatform/cloudsql-proxy) configured with the **connectionName** and **serviceAccount** parameters. If your business application is written in Go, you can use that library directly in your code. Otherwise, run it as a standalone container so your business application can reach the Cloud SQL instance by calling the proxy. In the official [tutorial](https://github.com/GoogleCloudPlatform/kubernetes-engine-samples/tree/master/service-catalog/cloud-sql-mysql), find the example on how to use the Cloud SQL Proxy with [Deployment](https://github.com/GoogleCloudPlatform/kubernetes-engine-samples/blob/master/service-catalog/cloud-sql-mysql/manifests/user-deployment.yaml#L46-L66). Unfortunately, the ServiceBindingUsage does not support the mounting volume functionality. To make it work, replace proxy container under the **spec.template.spec.containers** property with such entry:

```yaml
    - name: cloudsql-proxy
      image: gcr.io/cloudsql-docker/gce-proxy:1.11
      env:
      command: ["bin/sh"]
      args: ["-c", "echo $privateKeyData > /token; /cloud_sql_proxy -instances=$(connectionName)=tcp:3306 -credential_file=/token"]
```
