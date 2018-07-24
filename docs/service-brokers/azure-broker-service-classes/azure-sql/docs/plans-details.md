---
title: Services and Plans
type: Details
---

## Service description

The `azure-sql` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `basic` | "Basic Tier, 5 DTUs, 2GB, 7 days point-in-time restore |
| `standard-s0` | Standard Tier, 10 DTUs, 250GB, 35 days point-in-time restore |
| `standard-s1` | StandardS1 Tier, 20 DTUs, 250GB, 35 days point-in-time restore |
| `standard-s2` | StandardS2 Tier, 50 DTUs, 250GB, 35 days point-in-time restore |
| `standard-s3` | StandardS3 Tier, 100 DTUs, 250GB, 35 days point-in-time restore |
| `premium-p1` | PremiumP1 Tier, 125 DTUs, 500GB, 35 days point-in-time restore |
| `premium-p2` | PremiumP2 Tier, 250 DTUs, 500GB, 35 days point-in-time restore |
| `premium-p4` | PremiumP4 Tier, 500 DTUs, 500GB, 35 days point-in-time restore |
| `premium-p6` | PremiumP6 Tier, 1000 DTUs, 500GB, 35 days point-in-time restore |
| `premium-p11` | PremiumP11 Tier, 1750 DTUs, 1024GB, 35 days point-in-time restore |
| `data-warehouse-100` | DataWarehouse100 Tier, 100 DWUs, 1024GB |
| `data-warehouse-1200` | DataWarehouse1200 Tier, 1200 DWUs, 1024GB |

## Provision

This service provisions a new SQL DBMS and a new database upon that DBMS. The new
database is named randomly and is owned by a role (group) of the same name.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `Location` | `string` | The Azure region in which to provision applicable resources. | Y | None. |
| `Resource group"` | `string` | The new or existing resource group with which to associate new resources. | Y | Creates a new resource group with a UUID as its name. |
| `Firewall start IP address` | `string` | Specifies the start of the IP range that this firewall rule allows. | Y | `0.0.0.0` |
| `Firewall end IP address` | `string` | Specifies the end of the IP range that this firewall rule allows. | Y | `255.255.255.255` |

### Credentials

The binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| `host` | `string` | The fully-qualified address of the MySQL Server. |
| `port` | `int	` | The port number to connect to on the MySQL Server. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user. |
| `password` | `string` | The password for the database user. |
