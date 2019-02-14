---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Beta Plan` | Cloud Spanner plan for the Beta release of the Google Cloud Platform Service Broker |

## Provisioning parameters

Provisioning an instance creates a new Cloud Spanner instance. These are the input parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **config** | `string` | Determines where your Cloud Spanner data and Nodes are located. Affects cost, performance, and replication. These are the possible values: `nam-eur-asia1`, `nam3`, `regional-asia-east1`, `regional-asia-northeast1`, `regional-asia-south1`, `regional-europe-west1`, `regional-northamerica-northeast1`, `regional-us-central1`, `regional-us-east4`. This choice is permanent. | NO | `regional-us-central1` |
| **displayName** | `string` | Cloud Spanner display name. Must be 4-30 characters long. | YES | - |
| **instanceId** | `string` | Cloud Spanner unique and permanent identifier for instance. Use lowercase letters, numbers, or hyphens. Must be 6-30 characters long. | YES | - |
| **labels** | `object` | To organize your project, add arbitrary labels as key/value pairs to Cloud Spanner. Use labels to indicate different elements, such as Namespaces, services, or teams. | NO | - |
| **nodeCount** | `integer` | Number of Cloud Spanner Nodes. Add Nodes to increase data throughput and queries per second (QPS). Affects billing. Must contain minimum 1 Node. | YES | `1` |

## Update parameters

The update parameters are the same as the provisioning parameters.

## Binding parameters

Binding to an instance grants the provided service account access to the Cloud Spanner instance. Optionally, you can create a new service account and add the access to the Cloud Spanner instance. These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **createServiceAccount** | `boolean` | Creates a new service account for Spanner binding. | NO | `false` |
| **roles** | `array` | The list of Cloud Spanner roles for the binding. Affects the level of access granted to the service account. These are the possible values of this parameter: `roles/spanner.admin`, `roles/spanner.viewer`, `roles/spanner.databaseAdmin`, `roles/spanner.databaseUser`, `roles/spanner.databaseReader`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| **serviceAccount** | `string` | The GCP service account to which access is granted. | YES | - |

### Credentials

Binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **instanceId** | `string` | The ID of the instance. |
| **privateKeyData** | `JSON Object` | The service account OAuth information. |
| **projectId** | `string` | The ID of the project. |
| **serviceAccount** | `string` | The GCP service account to which access is granted. |
