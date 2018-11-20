---
title: Services and Plans
type: Details
---

## Service description

The `bigquery` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Beta Plan` | BigQuery plan for the Beta release of the Google Cloud Platform Service Broker |

## Provisioning parameters

This service creates a new BigQuery dataset when provisioning an instance. Binding to an instance grants the provided service account with access to the dataset or project. Optionally, a new service account can be created and given access to the Cloud Spanner instance. The provisioning parameters are as follows:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `datasetId` | `string` | A user-specified, unique ID for the BigQuery dataset. The minimal length of this value is 1 and the maximal is 1024. Must contain only letters, numbers, or underscores. | YES | - |
| `defaultTableExpirationMs` | `string` | The default lifetime of all tables in the dataset, in milliseconds. The minimum value is 3600000 milliseconds per one hour. Once this property is set, all newly-created tables in the dataset have an expirationTime property set to the creation time plus the value in this property, and changing the value only affects new tables, not existing ones. When a given table reaches the expirationTime, that table is deleted automatically. If you modify or remove the table's expirationTime before the table expires, or if you provide an explicit expirationTime when creating a table, that value takes precedence over the default expiration time indicated by this property. | NO | - |
| `description` | `string` | A user-friendly description of the BigQuery dataset. | NO | - |
| `friendlyName` | `string` | A descriptive name for the BigQuery dataset. | NO | - |
| `labels` | `object` | To organize your project, add arbitrary labels as key/value pairs to the BigQuery dataset. Use labels to indicate different elements, such as environments, services, or teams. | NO | - |
| `location` | `string` | The geographic location where the BigQuery dataset should reside. The value can be either `US` or `EU`. | NO | `US` |


## Update parameters:

These are the update parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `datasetId` | `string` | A user-specified, unique ID for the BigQuery dataset. The minimal length of this value is 1 and the maximal is 1024. Must contain only letters, numbers, or underscores. | YES | - |
| `defaultTableExpirationMs` | `string` | The default lifetime of all tables in the dataset, in milliseconds. The minimum value is 3600000 milliseconds per one hour. Once this property is set, all newly-created tables in the dataset have an expirationTime property set to the creation time plus the value in this property, and changing the value only affects new tables, not existing ones. When a given table reaches the expirationTime, that table is deleted automatically. If you modify or remove the table's expirationTime before the table expires, or if you provide an explicit expirationTime when creating a table, that value takes precedence over the default expiration time indicated by this property. | NO | - |
| `description` | `string` | A user-friendly description of the BigQuery dataset. | NO | - |
| `friendlyName` | `string` | A descriptive name for the BigQuery dataset. | NO | - |
| `labels` | `object` | To organize your project, add arbitrary labels as key/value pairs to the BigQuery dataset. Use labels to indicate different elements, such as environments, services, or teams. | NO | - |
| `location` | `string` | The geographic location where the BigQuery dataset should reside. The value can be either `US` or `EU`. | NO | `US` |


## Binding parameters:

These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `createServiceAccount` | `boolean` | Create a new service account for BigQuery binding. | NO | `false` |
| `roles` | `array` | The list of BigQuery roles for this binding. Affects the level of access granted to the service account. These are the possible values of this parameter: `roles/bigquery.dataOwner`, `roles/bigquery.dataEditor`, `roles/bigquery.dataViewer`, `roles/bigquery.user`, `roles/bigquery.jobUser`, `roles/bigquery.admin`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| `serviceAccount` | `string` | The GCP service account to which access is granted. | YES | - |
