---
title: Services and Plans
type: Details
---

## Service description

The `azure-eventhubs` service consist of the following plan:

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, 1 Consumer group, 100 Brokered connections |
| `standard` | Standard Tier, 20 Consumer groups, 1000 Brokered connections, Additional Storage, Publisher Policies |

## Provision

Provisions a new Event Hubs namespace and a new hub within it. The new namespace
and hub will be named using new UUIDs.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **location** | `string` | The Azure region in which to provision applicable resources. | Yes |  |
| **resourceGroup** | `string` | The (new or existing) resource group with which to associate new resources. | Yes |  |
| **tags** | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

## Bind

Returns a copy of one shared set of credentials.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| **connectionString** | `string` | Connection string. |
| **primaryKey** | `string` | Secret key (password). |
