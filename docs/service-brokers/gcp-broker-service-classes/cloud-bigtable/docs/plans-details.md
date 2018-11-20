---
title: Services and Plans
type: Details
---

## Service description

The `bigtable` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Beta Plan` | Bigtable plan for the Beta release of the Google Cloud Platform Service Broker |

## Provisioning parameters

Provisioning an instance creates a new Cloud Bigtable cluster and instance. Binding grants the provided service account with access on the Cloud Bigtable instance. Optionally, a new service account can be created and given access to the Cloud Bigtable instance. These are the input parameters to create a Bigtable instance:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `clusters` | `array of clusters` | Defines the cluster properties. The amount of items in the cluster must equal 1. A resizable group of nodes in a particular cloud location, capable of serving all Tables in the parent Instance. | YES | - |
| `displayName` | `string` | Cloud Bigtable display name. The minimal length of this value is 4 and the maximal is 30. | YES | - |
| `instanceId` | `string` |  Unique identifier for Cloud Bigtable instance. Permanent. Use lowercase letters, numbers, or hyphens. The minimal length of this value is 6 and the maximal is 33. | YES | - |
| `labels` | `object` | To organize your project, add arbitrary labels as key/value pairs to Cloud Bigtable. Use labels to indicate different elements, such as environments, services, or teams. | NO | - |
| `tables` | `array` | The tables present in the requested instance. A collection of user data indexed by row, column, and timestamp. Each table is served using the resources of its parent cluster. | NO | - |
| `type` | `string` | If your Cloud Bigtable cluster is meant to serve data to production, choose Production. If you want to experiment with Bigtable without committing to a production-grade cluster, choose Development. However, no Service Level Agreement (SLA) will apply. The value of this parameter can be either `PRODUCTION` or `DEVELOPMENT`. | NO | `PRODUCTION` |

###CLUSTER
| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `clusterId` | `string` | Unique identifier for Cloud Bigtable instance. Permanent. Use lowercase letters, numbers, or hyphens. The minimal length of this value is 6 and the maximal is 30. | YES | - |
| `defaultStorageType` | `string` | Storage type affects node performance and monthly storage costs. Choice is permanent.  The value of this parameter can be either `SSD` or `HDD`. | YES | `SSD` |
| `location` | `string` |  Choice is permanent. Determines where Cloud Bigtable data is stored. To reduce latency and increase throughput, store your data near the services that need it. These are the possible values of this parameter: `us-east1-b`,
`us-east1-c`, `asia-east1-b`, `asia-east1-a`, `asia-northeast1-c`, `asia-northeast1-b`, `europe-west1-b`, `europe-west1-c`, `europe-west4-b`, `europe-west1-d`, `us-central1-c`, `us-central1-b`, `us-central1-f`, `asia-southeast1-b`. | YES | - |
| `serveNodes` | `string` | Add nodes to increase capacity for data throughput and queries per second (QPS). Only applies to PRODUCTION. 3 nodes min. | NO | - |


###TABLES
| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `columnFamily` | `array` | The column families configured for this table, mapped by column family ID. Views: `SCHEMA_VIEW`, `FULL A set of columns within a table which share a common configuration. | NO | - |
| `ColumnFamily.columnFamilyId` | `object` | Cloud Bigtable display name. The minimal length of this value is 4 and the maximal is 30. | NO | - |
| `ColumnFamily.gcRule` | `object` |  Garbage collection rule specified as a protobuf. Must serialize to at most 500 bytes.\nNOTE: Garbage collection executes opportunistically in the background, and so it's possible for reads to return a cell even if it matches the active GC expression for its family. Rule for determining which cells to delete during garbage collection. | NO | - |
| `ColumnFamily.gcRule.maxAge` | `string` |  Delete cells in a column older than the given age. Values must be at least one millisecond, and will be truncated to microsecond granularity. | NO | - |
| `ColumnFamily.gcRule.maxNumVersions` | `integer` |  Delete all cells in a column except the most recent N. | NO | - |
| `granularity` | `string` |  The granularity (i.e. `MILLIS`) at which timestamps are stored in this table. Timestamps not matching the granularity will be rejected. If unspecified at creation time, the value will be set to `MILLIS`. Views: `SCHEMA_VIEW`, `FULL`\nMILLIS: The table keeps data versioned at a granularity of 1ms. The value of this parameter is `MILLIS`. | NO | `MILLIS` |
| `initialSplits` | `array` |  The optional list of row keys that will be used to initially split the\ntable into several tablets (tablets are similar to HBase regions. | NO | - |
| `initialSplits.key` | `string` |  "Row key to use as an initial tablet boundary. | NO | - |
| `tableId` | `string` |  The name by which the new table should be referred to within the parent instance, e.g., `foobar` rather than `\u003cparent\u003e/tables/foobar`. | YES | - |


## Update parameters:

These are the update parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `clusters` | `array of clusters` | Defines the cluster properties. The amount of items in the cluster must equal 1. A resizable group of nodes in a particular cloud location, capable of serving all Tables in the parent Instance. | YES | - |
| `displayName` | `string` | Cloud Bigtable display name. The minimal length of this value is 4 and the maximal is 30. | YES | - |
| `instanceId` | `string` |  Unique identifier for Cloud Bigtable instance. Permanent. Use lowercase letters, numbers, or hyphens. The minimal length of this value is 6 and the maximal is 33. | YES | - |
| `labels` | `object` | To organize your project, add arbitrary labels as key/value pairs to Cloud Bigtable. Use labels to indicate different elements, such as environments, services, or teams. | NO | - |
| `tables` | `array` | The tables present in the requested instance. A collection of user data indexed by row, column, and timestamp. Each table is served using the resources of its parent cluster. | NO | - |
| `type` | `string` | If your Cloud Bigtable cluster is meant to serve data to production, choose Production. If you want to experiment with Bigtable without committing to a production-grade cluster, choose Development. However, no Service Level Agreement (SLA) will apply. The value of this parameter can be either `PRODUCTION` or `DEVELOPMENT`. | NO | `PRODUCTION` |

###CLUSTER
| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `clusterId` | `string` | Unique identifier for Cloud Bigtable instance. Permanent. Use lowercase letters, numbers, or hyphens. The minimal length of this value is 6 and the maximal is 30. | YES | - |
| `defaultStorageType` | `string` | Storage type affects node performance and monthly storage costs. Choice is permanent.  The value of this parameter can be either `SSD` or `HDD`. | YES | `SSD` |
| `location` | `string` |  Choice is permanent. Determines where Cloud Bigtable data is stored. To reduce latency and increase throughput, store your data near the services that need it. These are the possible values of this parameter: `us-east1-b`,
`us-east1-c`, `asia-east1-b`, `asia-east1-a`, `asia-northeast1-c`, `asia-northeast1-b`, `europe-west1-b`, `europe-west1-c`, `europe-west4-b`, `europe-west1-d`, `us-central1-c`, `us-central1-b`, `us-central1-f`, `asia-southeast1-b`. | YES | - |
| `serveNodes` | `string` | Add nodes to increase capacity for data throughput and queries per second (QPS). Only applies to PRODUCTION. 3 nodes min. | NO | - |


###TABLES
| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `columnFamily` | `array` | The column families configured for this table, mapped by column family ID. Views: `SCHEMA_VIEW`, `FULL A set of columns within a table which share a common configuration. | NO | - |
| `ColumnFamily.columnFamilyId` | `object` | Cloud Bigtable display name. The minimal length of this value is 4 and the maximal is 30. | NO | - |
| `ColumnFamily.gcRule` | `object` |  Garbage collection rule specified as a protobuf. Must serialize to at most 500 bytes.\nNOTE: Garbage collection executes opportunistically in the background, and so it's possible for reads to return a cell even if it matches the active GC expression for its family. Rule for determining which cells to delete during garbage collection. | NO | - |
| `ColumnFamily.gcRule.maxAge` | `string` |  Delete cells in a column older than the given age. Values must be at least one millisecond, and will be truncated to microsecond granularity. | NO | - |
| `ColumnFamily.gcRule.maxNumVersions` | `integer` |  Delete all cells in a column except the most recent N. | NO | - |
| `granularity` | `string` |  The granularity (i.e. `MILLIS`) at which timestamps are stored in this table. Timestamps not matching the granularity will be rejected. If unspecified at creation time, the value will be set to `MILLIS`. Views: `SCHEMA_VIEW`, `FULL`\nMILLIS: The table keeps data versioned at a granularity of 1ms. The value of this parameter is `MILLIS`. | NO | `MILLIS` |
| `initialSplits` | `array` |  The optional list of row keys that will be used to initially split the\ntable into several tablets (tablets are similar to HBase regions. | NO | - |
| `initialSplits.key` | `string` |  "Row key to use as an initial tablet boundary. | NO | - |
| `tableId` | `string` |  The name by which the new table should be referred to within the parent instance, e.g., `foobar` rather than `\u003cparent\u003e/tables/foobar`. | YES | - |

## Binding parameters:

These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `createServiceAccount` | `boolean` | Create a new service account for Bigtable binding. | NO | `false` |
| `roles` | `array` | The list of Cloud Bigtable roles for this binding. Affects the level of access granted to the service account. These are the possible values of this parameter: `roles/bigtable.admin`, `roles/bigtable.user`, `roles/bigtable.reader`, `roles/bigtable.viewer`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| `serviceAccount` | `string` | The GCP service account to which access is granted. | YES | - |
