---
title: Services and Plans
type: Details
---

## Service description

The `azure-sql` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Basic Tier` | Basic Tier, 5 DTUs, 2GB, 7 days point-in-time restore |
| `Standard Tier` | Standard Tier, up to 3000 DTUs, with 250GB storage, 35 days point-in-time restore |
| `General Purpouse (preview)` | General Purpose Tier, up to 80 vCores, up to 440 GB Memory, up to 1 TB storage, 7 days point-in-time restore |
| `Business Critical (preview)` | Business Critical Tier, up to 80 vCores, up to 440 GB Memory, up to 1 TB storage, Local SSD, 7 days point-in-time restore. Offers highest resilience to failures using several isolated replicas |
| `Premium Tier` | Premium Tier, up to 4000 DTUs, with 500GB storage, 35 days point-in-time restore |

## Provision

This service provisions a new SQL DBMS and a new database upon that DBMS. The new
database is named randomly and is owned by a role (group) of the same name.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `Location` | `string` | The Azure region in which to provision applicable resources. | Y | None. |
| `Resource group` | `string` | The new or existing resource group with which to associate new resources. | Y | Creates a new resource group with a UUID as its name. |

### Credentials

The binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| `host` | `string` | The fully-qualified address of the MySQL Server. |
| `port` | `int	` | The port number to connect to on the MySQL Server. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user. |
| `password` | `string` | The password for the database user. |
