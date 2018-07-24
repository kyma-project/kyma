---
title: Overview
type: Overview
---

The Open Service Broker for Azure contains the **Azure SQL Database** service shown:

| Service Name | Description |
|--------------|-------------|
| `azure-sqldb` | Provision both an Azure SQL Server and a database upon that server. |

The `azure-sqldb` service allows you to provision both an SQL Server and a randomly named database. When the provisioning is successful, the database is ready to use. You cannot provision additional databases onto an instance provisioned through this service.

>**NOTE:** This version of the service is based on Open Service Broker for Azure, version 0.8.0-alpha, available on the [Azure](https://github.com/Azure/open-service-broker-azure/tree/v0.8.0-alpha) website.
For more information, see the [documentation](https://github.com/Azure/open-service-broker-azure/blob/v0.8.0-alpha/docs/modules/mssqldb.md).
