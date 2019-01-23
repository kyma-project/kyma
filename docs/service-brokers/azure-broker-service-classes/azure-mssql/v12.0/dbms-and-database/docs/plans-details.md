---
title: Services and Plans
type: Details
---

## Service description

The `azure-sql-12-0` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Basic Tier` | Basic Tier, 5 DTUs, 2GB, 7 days point-in-time restore |
| `Standard Tier` | Standard Tier, up to 3000 DTUs, with 250GB storage, 35 days point-in-time restore |
| `Premium Tier` | Premium Tier, up to 4000 DTUs, with 500GB storage, 35 days point-in-time restore |
| `General Purpouse (preview)` | General Purpose Tier, up to 80 vCores, up to 440 GB Memory, up to 1 TB storage, 7 days point-in-time restore |
| `Business Critical (preview)` | Business Critical Tier, up to 80 vCores, up to 440 GB Memory, up to 1 TB storage, Local SSD, 7 days point-in-time restore. Offers highest resilience to failures using several isolated replicas |

For applications which require less than 1 core, please use the basic, standard or premium plans.

## Provision

This service provisions a new SQL DBMS and a new database upon that DBMS. The new
database is named randomly and is owned by a role (group) of the same name.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **Location** | `string` | The Azure region in which to provision applicable resources. | Yes | None. |
| **Resource group** | `string` | The new or existing resource group with which to associate new resources. | Yes | Creates a new resource group with a UUID as its name. |
| **Tags** | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No | ags (even if none are specified) are automatically supplemented with heritage: open-service-broker-azure. |
| **Tags** | `array` | Tags to be applied to new resources, specified as key/value pairs. | No | ags (even if none are specified) are automatically supplemented with heritage: open-service-broker-azure. |
| **firewallRules**  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | No | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| **firewallRules[n].name** | `string` | Specifies the name of the generated firewall rule |Y | |
| **firewallRules[n].startIPAddress** | `string` | Specifies the start of the IP range allowed by this firewall rule | Yes | |
| **firewallRules[n].endIPAddress** | `string` | Specifies the end of the IP range allowed by this firewall rule | Yes | |
| **connectionPolicy** | `string` | Changes connection policy if you want. Refer to [here](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-connectivity-architecture#connection-policy). Valid values are "Redirect", "Proxy", and "Default". | No | |

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

Updates a previously provisioned SQL DB Database and DBMS.

### Updating Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **firewallRules**  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | No | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| **firewallRules[n].name** | `string` | Specifies the name of the generated firewall rule |Y | |
| **firewallRules[n].startIPAddress** | `string` | Specifies the start of the IP range allowed by this firewall rule | Yes | |
| **firewallRules[n].endIPAddress** | `string` | Specifies the end of the IP range allowed by this firewall rule | Yes | |
| **connectionPolicy** | `string` | Changes connection policy if you want. Refer to [here](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-connectivity-architecture#connection-policy). Valid values are "Redirect", "Proxy", and "Default". | No | |

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

Deletes both the database and the SQL Server instance.