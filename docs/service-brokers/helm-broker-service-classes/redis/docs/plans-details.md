---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `enterprise` |  |
| `micro` | Redis micro plan which uses the in-memory persistence. |


## Provision

This service provisions a new Redis cache.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `imagePullPolicy` | `string` | The possible values are `Always`, `IfNotPresent`, `Never`. | Y | `IfNotPresent` |


### Credentials

The binding returns the following credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **host** | `string` | The fully-qualified address of the Redis cache. |
| **port** | `int` | The port number to connect to the Redis cache. |
| **redis_password** | `string` | The password to the Redis cache. |
