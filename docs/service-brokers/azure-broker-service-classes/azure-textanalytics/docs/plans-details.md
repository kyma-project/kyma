---
title: Services and Plans
type: Details
---

## Service description

The `azure-textanalytics` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `free` | Free with 5,000 monthly transactions and no overage. |
| `standard-s0` | 25,000 monthly transactions and 3.00 per 1,000 overage. |
| `standard-s1` | 100,000 monthly transactions and 2.50 per 1,000 overage. |
| `standard-s2` | 500,000 monthly transactions and 2.00 per 1,000 overage. |
| `standard-s3` | 2,500,000 monthly transactions and 1.00 per 1,000 overage. |
| `standard-s4` | 10,000,000 monthly transactions and .50 per 1,000 overage. |

## Provision

Provisions a new text analytics API.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **location** | `string` | The Azure region in which to provision applicable resources. | Yes |  |
| **resourceGroup** | `string` | The (new or existing) resource group with which to associate new resources. | Yes |  |
| **tags** | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

### Credentials

Binding returns the following connection details:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **textAnalyticsEndpoint** | `string` | The text analytics API endpoint address. |
| **textAnalyticsKey** | `string` | The text analytics API access key. |
| **textAnalyticsName** | `string` | The name of the text analytics API. |

