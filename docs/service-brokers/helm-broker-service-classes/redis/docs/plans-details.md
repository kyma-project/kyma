---
title: Services and Plans
type: Details
---

## Service description

The `azure-rediscache` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `enterprise` | Basic Tier, 250MB Cache |
| `micro` | Standard Tier, 1GB Cache |


## Provision

This service provisions a new Redis cache.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `Labels` | `string` | The Azure region in which to provision applicable resources. | NO | - |
| `Image Pull Policy` | `string` | Possible values are `Always`, `IfNotPresent`, `Never`. | Y | `IfNotPresent` |

### Credentials

The binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| `host` | `string` | The fully-qualified address of the Redis cache. |
| `port` | `int	` | The port number to connect to on the Redis cache. |
| `password` | `string` | The password for the Redis cache. |
