---
title: Services and Plans
type: Details
---

## Service description

The `azure-sql-12-0-dr-database-pair-registered` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, 5 DTUs, 2GB, 7 days point-in-time restore |
| `standard` | Standard Tier, Up to 3000 DTUs, with 250GB storage, 35 days point-in-time restore |
| `premium` | Premium Tier, Up to 4000 DTUs, with 500GB storage, 35 days point-in-time restore |
| `general-purpose` | General Purpose Tier, Up to 80 vCores, Up to 440 GB Memory, Up to 1 TB storage, 7 days point-in-time restore |
| `business-critical` | Business Critical Tier, Up to 80 vCores, Up to 440 GB Memory, Up to 1 TB storage, Local SSD, 7 days point-in-time restore. Offers highest resilience to failures using several isolated replicas |

## Provision

Validates the database specified in the provisioning parameter.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **parentAlias** | `string` | Specifies the alias of the DBMS upon which the database should be provisioned. | Yes | |
| **database** | `string` | Specifies the name of the databases. | Yes | |
| **failoverGroup** | `string` | Specifies the name of the failover group. | Yes | |

Additional Provision Parameters for : standard plan

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **dtus** | `integer` | Specifies Database transaction units, which represent a bundled measure of compute, storage, and IO resources. Valid values are 10, 20, 50, 100, 200, 400, 800, 1600, 3000 | No | 10 |


Additional Provision Parameters for : premium plan

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **dtus** | `integer` | Specifies Database transaction units, which represent a bundled measure of compute, storage, and IO resources. Valid values are 125, 250, 500, 1000, 1750, 4000 | No | 125 |

Additional Provision Parameters for: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 24, 32, 48, 80 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | No | 5 |

Additional Provision Parameters for: business-critical

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 24, 32, 48, 80 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | No | 5 |

## Update

Update operation is not supported.