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
| **accountId** | `string` | A user-specified ID, which is the unique name of a GCP Service Account. Must start with a lower case letter, followed by lower case alphanumeric characters separated by hyphens. The minimal length of this value is 6 and the maximal is 30. | YES | - |
| **displayName** | `string` | Optionally add a descriptive name of the Service Account. The maximal length is 100. | NO | - |

## Update parameters:

The update parameters are the same as the provisioning parameters.

## Binding

Binding makes the Cloud IAM service account private key available to your application.
