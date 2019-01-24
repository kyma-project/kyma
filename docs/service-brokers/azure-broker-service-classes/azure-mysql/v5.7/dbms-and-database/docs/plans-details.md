---
title: Services and Plans
type: Details
---

## Services & Plans

Service class contains the following plans and parameters:

| Plan Name | Description |
|-----------|-------------|
| `Basic Tier` | Basic Tier, up to 2 vCores, variable I/O performance |
| `General Purpose Tier` | General Purpose Tier, up to 32 vCores, predictable I/O Performance, local or geo-redundant backups |
| `Memory Optimized Tier` | Memory Optimized Tier, up to 16 memory optimized vCores, predictable I/O Performance, local or geo-redundant backups |

## Provision

Provisions a new MySQL DBMS and a new database upon it. The new database will be named randomly.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **location** | `string` | The Azure region in which to provision applicable resources. | Yes | |
| **resourceGroup** | `string` | The (new or existing) resource group with which to associate new resources. | Yes | |
| **sslEnforcement** | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | No | `""`. Left unspecified, SSL _will_ be enforced. |
| **firewallRules**  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | No | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| **firewallRules[n].name** | `string` | Specifies the name of the generated firewall rule |Yes | |
| **firewallRules[n].startIPAddress** | `string` | Specifies the start of the IP range allowed by this firewall rule | Yes | |
| **firewallRules[n].endIPAddress** | `string` | Specifies the end of the IP range allowed by this firewall rule | Yes | |
| **tags** | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

The three plans each have additional provisioning parameters with different default and allowed values. See the tables below for details on each.

#### Provisioning Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | No | 1 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |

#### Provisioning Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16 or 32 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 2048 | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |
| **backupRedundancy** | `string` | Specifies the backup redundancy, either `local` or `geo` | No | `local` |

#### Provisioning Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8 or 16 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 2048 | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |
| **backupRedundancy** | `string` | Specifies the backup redundancy, either `local` or `geo` | No | `local` |

## Bind

Creates a new user on the MySQL DBMS. The new user will be named randomly and
will be granted a wide array of permissions on the database.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| **host** | `string` | The fully-qualified address of the MySQL DBMS. |
| **port** | `int` | The port number to connect to on the MySQL DBMS. |
| **database** | `string` | The name of the database. |
| **username** | `string` | The name of the database user (in the form username@host). |
| **password** | `string` | The password for the database user. |
| **sslRequired** | `boolean` | Flag indicating if SSL is required to connect the MySQL DBMS. |
| **uri** | `string` | A URI string containing all necessary connection information. |
| **tags** | `string[]` | A list of tags consumers can use to identify the credential. |
