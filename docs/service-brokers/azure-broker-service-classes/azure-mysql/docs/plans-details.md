---
title: Services and Plans
type: Details
---

## Service description
This service is named `azure-mysqldb` with the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `MYSQLB50` | Basic Tier, 50 DTUs |
| `MYSQLB100` | Basic Tier, 100 DTUs |
| `MYSQLS100` | Standard Tier, 100 DTUs |
| `MYSQLS200` | Standard Tier, 200 DTUs |
| `MYSQLS400` | Standard Tier, 400 DTUs |
| `MYSQLS800` | Standard Tier, 800 DTUs |

## Provision

This service provisions a new MySQL DBMS and a new database upon it. The new database is named randomly.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `Location` | `string` | The Azure region in which to provision applicable resources. | Y | None. |
| `Resource group` | `string` | The (new or existing) resource group with which to associate new resources. | Y | Creates a new resource group with a UUID as its name. |
| `Firewall start IP address` | `string` | Specifies the start of the IP range that the firewall rule allows. | Y | `0.0.0.0` |
| `Firewall end IP address` | `string` | Specifies the end of the IP range that the firewall rule allows. | Y | `255.255.255.255` |

### Credentials

The binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| `host` | `string` | The fully-qualified address of the SQL Server. |
| `port` | `int	` | The port number to connect to on the SQL Server. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user. |
| `password` | `string` | The password for the database user. |
