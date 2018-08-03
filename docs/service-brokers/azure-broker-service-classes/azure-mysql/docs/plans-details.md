---
title: Services and Plans
type: Details
---

## Service description
This service is named `azure-mysqldb` with the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Basic Tier` | Basic Tier, Up to 2 vCores, Variable I/O performance |
| `General Purpose Tier` | General Purporse Tier, Up to 32 vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |
| `Memory Optimized Tier` | Memory Optimized Tier, Up to 16 memory optimized vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |

## Provision

This service provisions a new MySQL DBMS and a new database upon it. The new database is named randomly.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `Backup redundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |
| `Backup retention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `Cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16 or 32 | N | 2 |
| `Location` | `string` | The Azure region in which to provision applicable resources. | Y | None. |
| `Resource group` | `string` | The (new or existing) resource group with which to associate new resources. | Y | Creates a new resource group with a UUID as its name. |
| `SSL Eeforcement` | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | `""`. Left unspecified, SSL _will_ be enforced. |
| `Storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |

### Credentials

The binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| `host` | `string` | The fully-qualified address of the SQL Server. |
| `port` | `int	` | The port number to connect to on the SQL Server. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user. |
| `password` | `string` | The password for the database user. |
