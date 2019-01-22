---
title: Overview
type: Overview
---

The Open Service Broker for Azure contains the **Azure SQL Database - database from existing** service shown:

| Service Name | Description |
|--------------|-------------|
| `azure-sql-12-0-database-from-existing` | Used to create SQL database service instance from existing Azure SQL Database *for taking over the database*. Typically, you can create **azure-sql-12-0-dbms-registered** service instance to register your Azure SQL server first and use this service to import the database to OSBA's management. |

>**NOTE:** This version of the service is based on Open Service Broker for Azure, version [1.3.1](https://github.com/Azure/open-service-broker-azure/releases).
For more information, see the [documentation](https://github.com/Azure/open-service-broker-azure/blob/v1.3.1/docs/modules/mssql.md).
