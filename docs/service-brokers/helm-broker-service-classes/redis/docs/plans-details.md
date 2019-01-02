---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

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
| **labels** | `object` | To organize your project, add arbitrary labels. Use them to indicate different elements, such as environments, services, or teams. | NO | - |
| **imagePullPolicy** | `string` | Specifies how the kubelet pulls images from the specified registry. The possible values are `Always` (the image is pulled every time the Pod is started) `IfNotPresent` (the image is pulled only if it is not already present locally), `Never` (the image is assumed to exist locally and there is no attempt to pull it). | YES | `IfNotPresent` |


### Credentials

The binding returns the following credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **host** | `string` | The fully-qualified address of the Redis cache. |
| **port** | `int` | The port number to connect to the Redis cache. |
| **redis_password** | `string` | The password to the Redis cache. |
