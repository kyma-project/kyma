---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Beta Plan` | Bigtable plan for the Beta release of the Google Cloud Platform Service Broker |

## Provisioning parameters

Provisioning an instance creates a new Cloud Bigtable cluster and instance. These are the input parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **clusters** | `array` | Defines the cluster properties. The amount of items in the cluster must equal 1. For more information, see the **Cluster properties** section. | YES | - |
| **displayName** | `string` | Cloud Bigtable display name. Must be 4-30 characters long. | YES | - |
| **instanceId** | `string` |  Unique and permanent identifier for the Cloud Bigtable instance. Use only lowercase letters, numbers, or hyphens. Must be 6-33 characters long. | YES | - |
| **labels** | `object` | To organize your project, add arbitrary labels as key/value pairs to Cloud Bigtable. Use labels to indicate different elements, such as Namespaces, services, or teams. | NO | - |
| **tables** | `array` | The tables present in the requested instance. Each table is served using the resources of its parent cluster. For more information, see the **Tables properties** section. | NO | - |
| **type** | `string` | The value of this parameter can be either `PRODUCTION` or `DEVELOPMENT`. If your Cloud Bigtable cluster serves data to production, choose `Production`. If you want to experiment with Bigtable without committing to a production-grade cluster, choose `Development`. However, no Service Level Agreement (SLA) applies. | NO | `PRODUCTION` |

### Cluster properties

These are the properties that you can set for your Cloud Bigtable cluster:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **clusterId** | `string` | Unique and permanent identifier for Cloud Bigtable instance. Use only lowercase letters, numbers, or hyphens. Must be 6-30 characters long. | YES | - |
| **defaultStorageType** | `string` | Storage type affects Node performance and monthly storage costs. The value of this parameter can be either `SSD` or `HDD`. The choice is permanent. | YES | `SSD` |
| **location** | `string` |  Determines where Cloud Bigtable data is stored. To reduce latency and increase throughput, store your data near the services that need it. These are the possible values: `us-east1-b`, `us-east1-c`, `asia-east1-b`, `asia-east1-a`, `asia-northeast1-c`, `asia-northeast1-b`, `europe-west1-b`, `europe-west1-c`, `europe-west4-b`, `europe-west1-d`, `us-central1-c`, `us-central1-b`, `us-central1-f`, `asia-southeast1-b`. The choice is permanent.  | YES | - |
| **serveNodes** | `string` | Add Nodes to increase capacity for data throughput and queries per second (QPS). Only applies to `PRODUCTION`. The minimal number of Nodes is 3. | NO | - |


### Tables properties

These are the **Tables** properties:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **tableId** | `string` |  The name by which you refer to the new table within the parent instance. | YES | - |
| **columnFamily** | `array` | A set of columns within a table which share a common configuration. | NO | - |
| **ColumnFamily.columnFamilyId** | `string` | The name of the ColumnFamily. | NO | - |
| **ColumnFamily.gcRule** | `object` |  Rule used to determine which cells to delete during garbage collection. Must serialize to at most 500 bytes. | NO | - |
| **ColumnFamily.gcRule.maxAge** | `string` |  Deletes cells in a column older than a given age. Values must be at least one millisecond, and are truncated to microsecond granularity. | NO | - |
| **ColumnFamily.gcRule.maxNumVersions** | `integer` |  Deletes all cells in a column, except the most recent. | NO | - |
| **granularity** | `string` |  The granularity at which timestamps are stored in this table. Timestamps not matching the granularity are rejected. | NO | `MILLIS` |
| **initialSplits** | `array` |  The optional list of row keys that are used to initially split the table into several tablets. | NO | - |
| **initialSplits.key** | `string` |  Row key to use as an initial tablet boundary. | NO | - |


## Update parameters

The update parameters are the same as the provisioning parameters.

## Binding parameters

Binding grants the provided service account access to the Cloud Bigtable instance. Optionally, you can create a new service account and add the access to the Cloud Bigtable instance. These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **createServiceAccount** | `boolean` | Creates a new service account for Bigtable binding. | NO | `false` |
| **roles** | `array` | The list of Cloud Bigtable roles for the binding. Affects the level of access granted to the service account. These are the possible values: `roles/bigtable.admin`, `roles/bigtable.user`, `roles/bigtable.reader`, `roles/bigtable.viewer`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| **serviceAccount** | `string` | The GCP service account to which access is granted. | YES | - |

### Credentials

Binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **instanceId** | `string` | The ID of the instance. |
| **privateKeyData** | `JSON Object` | The service account OAuth information. |
| **projectId** | `string` | The ID of the project. |
| **serviceAccount** | `string` | The GCP service account to which access is granted. |
