---
title: Services and Plans
type: Details
---

## Service description

The `azure-postgresql-9.6-dbms` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, Up to 2 vCores, Variable I/O performance |
| `general-purpose` | General Purporse Tier, Up to 32 vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |
| `memory-optimized` | Memory Optimized Tier, Up to 16 memory optimized vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |

## Provision

Provisions an Azure Database for PostgreSQL DBMS instance containing no databases. Databases can be created through subsequent provision requests using the `azure-postgresql-database` service.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **location** | `string` | The Azure region in which to provision applicable resources. | Yes | |
| **resourceGroup** | `string` | The (new or existing) resource group with which to associate new resources. | Yes | |
| **alias** | `string` | Specifies an alias that can be used by later provision actions to create databases on this DBMS. | Yes | |
| **sslEnforcement** | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | No | `""`. Left unspecified, SSL _will_ be enforced. |
| **firewallRules**  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | No | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| **firewallRules[n].name** | `string` | Specifies the name of the generated firewall rule |Yes | |
| **firewallRules[n].startIPAddress** | `string` | Specifies the start of the IP range allowed by this firewall rule | Yes | |
| **firewallRules[n].endIPAddress** | `string` | Specifies the end of the IP range allowed by this firewall rule | Yes | |
| **virtualNetworkRules**  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | No | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| **virtualNetworkRules[n].name** | `string` | Specifies the name of the generated virtual network rule |Yes | |
| **virtualNetworkRules[n].subnetId** | `string` | The full resource ID of a subnet in a virtual network to allow access from. Example format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vn}/subnets/{sn} | Yes | |
| **tags** | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |
| **extensions** | `string[]` | Specifies a list of PostgreSQL extensions to install | No | |

#### Provisioning Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | No | 1 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |


#### Provisioning Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 32 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |
| **backupRedundancy** | `string` | Specifies the backup redundancy, either `local` or `geo` | No | `local` |

#### Provisioning Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, or 16 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |
| **backupRedundancy** | `string` | Specifies the backup redundancy, either `local` or `geo` | No | `local` |

## Update

Updates a previously provisioned PostgreSQL DBMS.

### Updating Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **sslEnforcement** | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | No | `""`. Left unspecified, SSL _will_ be enforced. |
| **firewallRules**  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | No | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| **firewallRules[n].name** | `string` | Specifies the name of the generated firewall rule |Y | |
| **firewallRules[n].startIPAddress** | `string` | Specifies the start of the IP range allowed by this firewall rule | Yes | |
| **firewallRules[n].endIPAddress** | `string` | Specifies the end of the IP range allowed by this firewall rule | Yes | |

#### Updating Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | No | 1 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |


#### Updating Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 32 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |

#### Updating Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **cores** | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, or 16 | No | 2 |
| **storage** | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | No | 10 |
| **backupRetention** | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | No | 7 |

## Deprovision

Deletes the PostgreSQL DBMS only. If databases have been provisioned on this DBMS, deprovisioning will be deferred until all databases have been deprovisioned.
