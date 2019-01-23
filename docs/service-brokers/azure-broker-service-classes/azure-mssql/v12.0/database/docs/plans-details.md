---
title: Services and Plans
type: Details
---

## Service description

The `azure-sql-12-0-database` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, 5 DTUs, 2GB, 7 days point-in-time restore |
| `standard` | Standard Tier, Up to 3000 DTUs, with 250GB storage, 35 days point-in-time restore |
| `premium` | Premium Tier, Up to 4000 DTUs, with 500GB storage, 35 days point-in-time restore |
| `general-purpose` | General Purpose Tier, Up to 80 vCores, Up to 440 GB Memory, Up to 1 TB storage, 7 days point-in-time restore |
| `business-critical` | Business Critical Tier, Up to 80 vCores, Up to 440 GB Memory, Up to 1 TB storage, Local SSD, 7 days point-in-time restore. Offers highest resilience to failures using several isolated replicas |

## Provision

Provisions a new database upon an existing server. The new database will be named randomly. If the DBMS does not yet exist, provision of the database will be deferred until the DBMS has been provisioned.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **parentAlias** | `string` | Specifies the alias of the DBMS upon which the database should be provisioned. | Yes | |

Additional Provision Parameters for : standard plan

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **dtus** | `integer` | Specifies Database transaction units, which represent a bundled measure of compute, storage, and IO resources. Valid values are 10, 20, 50, 100, 200, 400, 800, 1600, 3000 | No | 10 |


Additional Provision Parameters for : premium plan

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **dtus** | `integer` | Specifies Database transaction units, which represent a bundled measure of compute, storage, and IO resources. Valid values are 125, 250, 500, 1000, 1750, 1000 | No | 125 |

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

## Bind

Creates a new user on the SQL Server. The new user will be named randomly and granted permission to log into and administer the database.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| **host** | `string` | The fully-qualified address of the SQL Server. |
| **port** | `int` | The port number to connect to on the SQL Server. |
| **database** | `string` | The name of the database. |
| **username** | `string` | The name of the database user (in the form username@host). |
| **password** | `string` | The password for the database user. |
| **uri** | `string` | A uri string containing connection information. |
| **jdbcUrl** | `string` | A fully formed JDBC url. |
| **encrypt** | `boolean` | Flag indicating if the connection should be encrypted. |
| **tags** | `string[]` | List of tags. |

## Update

Updates a previously provisioned SQL DB Database.

### Updating Parameters

Parameters for updating the SQL DB Database differ by plan. See each section for relevant parameters.

Additional Provision Parameters for : standard plan

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **dtu** | `integer` | Specifies Database transaction units, which represent a bundled measure of compute, storage, and IO resources. Valid values are 10, 20, 50, 100, 200, 400, 800, 1600, 3000 | No | 10 |

Additional Provision Parameters for : premium plan

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **dtu** | `integer` | Specifies Database transaction units, which represent a bundled measure of compute, storage, and IO resources. Valid values are 125, 250, 500, 1000, 1750, 1000 | No | 125 |

Additional Provision Parameters for: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 24, 32, 48, 80 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, decreasing storage is not currently supported | No | 5 |

Additional Provision Parameters for: business-critical

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 24, 32, 48, 80 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, decreasing storage is not currently supported | No | 5 |

## Unbind

Drops the applicable user from the SQL Server.

## Deprovision

Deletes the database from the SQL Server instance, but does not delete the DBMS.
