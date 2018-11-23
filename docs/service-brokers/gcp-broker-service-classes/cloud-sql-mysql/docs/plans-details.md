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
| `databaseVersion` | `string` | The database engine type and version. The value can be either `MYSQL_5_7` or `MYSQL_5_6`. The choice is permanent. | NO | `MYSQL_5_7` |
| `failoverReplica` | `object` | The name and status of the failover replica. This property is applicable only to Second Generation instances. | NO | - |
| `instanceId` | `string` | CloudSQL instance ID. Use lowercase letters, numbers, and hyphens. Start with a letter. Choice is permanent.| YES | - |
| `onPremisesConfiguration` | `object` | Configuration specific to on-premises instances.| NO | - |
| `region` | `string` | Determines where your CloudSQL data is located. For better performance, keep your data close to the services that need it. Choice is permanet.| NO | `us-central1` |
| `replicaConfiguration` | `object` | Configuration specific to read-replicas replicating from on-premises masters. | NO | - |
| `settings` | `object` | The user settings. | YES | - |

## Update parameters:

The update parameters are the same as the provisioning parameters.

## Binding parameters:

Binding grants one of available IAM roles on the Cloud SQL instance to the specified service account. Optionally, a new service account can be created and given access to the MySQL instance. These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `createServiceAccount` | `boolean` | Creates a new service account for Spanner binding. | NO | `false` |
| `roles` | `array` | The list of Cloud Spanner roles for the binding. Affects the level of access granted to the service account. These are the possible values of this parameter: `roles/spanner.admin`, `roles/spanner.viewer`, `roles/spanner.databaseAdmin`, `roles/spanner.databaseUser`, `roles/spanner.databaseReader`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| `serviceAccount` | `string` | The GCP service account to which access is granted. | YES | - |
