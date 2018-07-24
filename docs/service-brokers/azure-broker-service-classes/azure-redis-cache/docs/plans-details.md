---
title: Services and Plans
type: Details
---

## Service description

The `azure-rediscache` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, 250MB Cache |
| `standard` | Standard Tier, 1GB Cache |
| `premium` | Premium Tier, 6GB Cache |

## Provision

This service provisions a new Redis cache.

### Provisioning parameters
These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `Location` | `string` | The Azure region in which to provision applicable resources. | Y | None. |
| `Resource group` | `string` | The new or existing resource group with which to associate new resources. | Y | Creates a new resource group with a UUID as its name. |
| `Server name` | `string` | The name of the Azure Redis Cache to create. | N |  |

### Credentials

The binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| `host` | `string` | The fully-qualified address of the Redis cache. |
| `port` | `int	` | The port number to connect to on the Redis cache. |
| `password` | `string` | The password for the Redis cache. |
