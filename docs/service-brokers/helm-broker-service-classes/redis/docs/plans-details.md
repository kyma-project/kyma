---
title: Services and Plans
type: Details
---

## Service description

The Redis Service Class provides the following plans:

| Plan Name | Description |
|-----------|-------------|
| `enterprise` | Redis enterprise plan which uses Persistent Volume Claim (PVC). |
| `micro` | Redis micro plan which uses the in-memory persistence. |


## Provision

This service provisions a new Redis cache.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **imagePullPolicy** | `string` | Specifies how the kubelet pulls images from the specified registry. The possible values are `Always`, `IfNotPresent`, `Never`. | NO | `IfNotPresent` |

## Credentials

The binding creates a Secret with the following credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **HOST** | `string` | The fully-qualified address of the Redis cache. |
| **PORT** | `int` | The port number to connect to the Redis cache. |
| **REDIS_PASSWORD** | `string` | The password to the Redis cache. |
