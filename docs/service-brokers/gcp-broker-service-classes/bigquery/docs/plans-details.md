---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Beta Plan` | BigQuery plan for the Beta release of the Google Cloud Platform Service Broker |

## Provisioning parameters

Provisioning an instance creates a new BigQuery dataset. The provisioning parameters are as follows:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **datasetId** | `string` | A user-specified, unique ID for the BigQuery dataset. Must be 1-1024 characters long. Must contain only letters, numbers, or underscores. | YES | - |
| **defaultTableExpirationMs** | `string` | The default lifetime of all tables in the dataset in milliseconds. The minimum value is `3600000` milliseconds per one hour. Once this property is set, all newly-created tables in the dataset have the **expirationTime** property set to the creation time plus the value in the **defaultTableExpirationMs** parameter. Changing the value only affects new tables, not existing ones. When a given table reaches the **expirationTime**, that table is deleted automatically. If you modify or remove the table's **expirationTime** before the table expires, or if you provide an explicit **expirationTime** when creating a table, that value takes precedence over the default expiration time indicated by this property. | NO |  `3600000` |
| **description** | `string` | A user-friendly description of the BigQuery dataset. | NO | - |
| **friendlyName** | `string` | A descriptive name for the BigQuery dataset. | NO | - |
| **labels** | `object` | To organize your project, add arbitrary labels as key/value pairs to the BigQuery dataset. Use labels to indicate different elements, such as Namespaces, services, or teams. | NO | - |
| **location** | `string` | The geographic location where the BigQuery dataset resides. The value can be either `US` or `EU`. | NO | `US` |


## Update parameters

The update parameters are the same as the provisioning parameters.

## Binding parameters

Binding to an instance grants the provided service account the access to the dataset or project. Optionally, you can create a new service account and add the access to the Cloud Spanner instance. These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **createServiceAccount** | `boolean` | Create a new service account for BigQuery binding. | NO | `false` |
| **roles** | `array` | The list of BigQuery roles for the binding. Affects the level of access granted to the service account. These are the possible values: `roles/bigquery.dataOwner`, `roles/bigquery.dataEditor`, `roles/bigquery.dataViewer`, `roles/bigquery.user`, `roles/bigquery.jobUser`, `roles/bigquery.admin`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| **serviceAccount** | `string` | The GCP service account to which access is granted. | YES | - |

### Credentials

Binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **datasetId** | `string` | The ID of the dataset. |
| **privateKeyData** | `JSON Object` | The service account OAuth information. |
| **projectId** | `string` | The ID of the project. |
| **serviceAccount** | `string` | The GCP service account to which access is granted. |
