---
title: Overview
type: Overview
---

The Open Service Broker for Azure contains the **Azure SQL Database Failover Group** service shown:

| Service Name | Description |
|--------------|-------------|
| `azure-sql-12-0-dr-database-pair-registered` | Used to register existing **azure-sql-12-0-dr-database-pair** as a service instance. It is for your OSBA instances in other regions to use the same failover group. It doesn't create new databases and doesn't delete databases but only validates the databases. |

The service requires `ENABLE_DISASTER_RECOVERY_SERVICES` to be `true` in OSBA environment variables.

>**NOTE:** This version of the service is based on [Open Service Broker for Azure](https://github.com/Azure/open-service-broker-azure).
For more details, read the **Plans Details** document.
