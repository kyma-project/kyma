---
title: Services and Plans
type: Details
---

## Service description

The `azure-keyvault` service consist of the following plan:

| Plan Name | Description |
|-----------|-------------|
| `standard` | Standard Tier |
| `premium` | Premium Tier |

## Provision

Provisions a new Key Vault. The new vault will be named using a new UUID.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **clientId** | `string` | Client ID (username) for an existing service principal, which will be granted access to the new vault.| Yes | |
| **clientSecret** | `string` | Client secret (password) for an existing service principal, which will be granted access to the new vault. __WARNING: This secret will be shared with all users who bind to the vault!__ | Yes | |
| **location** | `string` | The Azure region in which to provision applicable resources. | Yes |  |
| **objectid** | `string` | Object ID for an existing service principal, which will be granted access to the new vault. | Yes | |
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
| **vaultUri** | `string` | Fully qualified URI for connecting to the vault. |
| **clientId** | `string` | Service principal client ID (username) to use when connecting to the vault. |
| **clientSecret** | `string` | Service principal client secret (password) to use when connecting to the vault. |

