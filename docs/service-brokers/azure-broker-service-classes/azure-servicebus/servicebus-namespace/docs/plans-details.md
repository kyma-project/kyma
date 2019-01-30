---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, Shared Capacity |
| `standard` | Standard Tier, Shared Capacity, Topics, 12.5M Messaging Operations/Month, Variable Pricing |
| `premium` | Premium Tier, Dedicated Capacity, Recommended For Production Workloads, Fixed Pricing |

## Provision

Provisions a new Service Bus namespace. The new namespace will be named using
new UUIDs.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **location** | `string` | The Azure region in which to provision applicable resources. | Yes |  |
| **resourceGroup** | `string` | The (new or existing) resource group with which to associate new resources. | Yes |  |
| **alias** | `string` | Specifies an alias that can be used by later provision actions to create queues/topics in this namespace. | Yes |  |
| **tags** | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| **connectionString** | `string` | Connection string. |
| **primaryKey** | `string` | Secret key (password). |
| **namespaceName** | `string` | The name of the namespace. |