---
title: Services and Plans
type: Details
---

## Services & Plans

Service class contains the following plans and parameters:

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, Up to 2 vCores, Variable I/O performance |
| `general-purpose` | General Purporse Tier, Up to 32 vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |
| `memory-optimized` | Memory Optimized Tier, Up to 16 memory optimized vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |

## Provision

Provisions an Azure Database for MySQL DBMS instance containing no databases. Databases can be created through subsequent provision requests using the `azure-mysql-database` service.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `location` | `string` | The Azure region in which to provision applicable resources. | Y | |
| `resourceGroup` | `string` | The (new or existing) resource group with which to associate new resources. | Y | |
| `alias` | `string` | Specifies an alias that can be used by later provision actions to create databases on this DBMS. | Y | |
| `sslEnforcement` | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | `""`. Left unspecified, SSL _will_ be enforced. |
| `firewallRules`  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | N | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| `firewallRules[n].name` | `string` | Specifies the name of the generated firewall rule |Y | |
| `firewallRules[n].startIPAddress` | `string` | Specifies the start of the IP range allowed by this firewall rule | Y | |
| `firewallRules[n].endIPAddress` | `string` | Specifies the end of the IP range allowed by this firewall rule | Y | |
| `tags` | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | N | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

The three plans each have additional provisioning parameters with different default and allowed values. See the tables below for details on each.

#### Provisioning Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | N | 1 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |

#### Provisioning Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16 or 32 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 2048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `backupRedundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |

#### Provisioning Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8 or 16 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 2048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `backupRedundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |
