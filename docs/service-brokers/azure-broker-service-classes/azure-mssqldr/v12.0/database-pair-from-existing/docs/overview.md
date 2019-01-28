---
title: Overview
type: Overview
---

The Open Service Broker for Azure contains the **Azure SQL Database Failover Group** service shown:

| Service Name | Description |
|--------------|-------------|
| `azure-sql-12-0-dr-database-pair-from-existing` | Taking over an existing failover group (included the databases) upon a previous DBMS pair as a service instance. |

The service requires `ENABLE_DISASTER_RECOVERY_SERVICES` and `ENABLE_MIGRATION_SERVICES` to be `true` in OSBA environment variables.

>**NOTE:** This version of the service is based on [Open Service Broker for Azure](https://github.com/Azure/open-service-broker-azure).
For more details, read the **Plans Details** document.
