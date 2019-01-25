---
title: Overview
type: Overview
---

*Note: PostgreSQL version 9.6 is STABLE, and PostgreSQL version 10 is in PREVIEW*

The Open Service Broker for Azure contains the **Azure PostgreSQL Database** service shown:

| Service Type                  | Description                                                  |
| ----------------------------- | ------------------------------------------------------------ |
| `azure-postgresql-10-database`          | Provision both an Azure Database for PostgreSQL Database Management System (DBMS) and a database. |

The `azure-postgresql-10-database` service allow you to provision a database. When the provision operation is successful, the database will be ready to use. You can not provision additional databases onto an instance provisioned through these two services. The `azure-postgresql-10-dbms` and `azure-postgresql-10-database` services, on the other hand, can be combined to provision multiple databases on a single DBMS.

>**NOTE:** This version of the service is based on [Open Service Broker for Azure](https://github.com/Azure/open-service-broker-azure).
For more details, read the **Plans Details** document.
