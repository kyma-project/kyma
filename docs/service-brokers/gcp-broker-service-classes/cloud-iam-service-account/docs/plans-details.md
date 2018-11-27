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

Provisioning an instance creates a new Cloud IAM Service Account. These are the input parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **accountId** | `string` | A user-specified ID, which is the unique name of a GCP Service Account. Must start with a lower case letter, followed by lower case alphanumeric characters separated by hyphens. Must be 6-30 characters long. | YES | - |
| **displayName** | `string` | Optionally add a descriptive name of the Service Account. The maximal length is 100. | NO | - |

## Update parameters

The update parameters are the same as the provisioning parameters.

## Binding

Binding makes the Cloud IAM service account private key available to your application.

### Credentials

Binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **privateKeyData** | `JSON Object` | The service account OAuth information. |
| **serviceAccount** | `string` | The GCP service account to which access is granted. |
