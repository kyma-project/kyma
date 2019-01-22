---
title: Overview
type: Overview
---

The Open Service Broker for Azure contains the **Azure SQL Database Failover Group** service shown:

| Service Name | Description |
|--------------|-------------|
| `azure-sql-12-0-dr-database-pair-from-existing-primary` | Taking over an existing database upon the primary server of a previous DBMS pair, and extend it to a failover group deployment. |

The service requires `ENABLE_DISASTER_RECOVERY_SERVICES` and `ENABLE_MIGRATION_SERVICES` to be `true` in OSBA environment variables.

>**NOTE:** This version of the service is based on Open Service Broker for Azure, version [1.3.1](https://github.com/Azure/open-service-broker-azure/releases).
For more information, see the [documentation](https://github.com/Azure/open-service-broker-azure/blob/v1.3.1/docs/modules/mssqldr.md).
