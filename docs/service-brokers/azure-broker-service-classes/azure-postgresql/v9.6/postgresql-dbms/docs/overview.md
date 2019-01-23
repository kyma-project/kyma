---
title: Overview
type: Overview
---

*Note: PostgreSQL version 9.6 is STABLE, and PostgreSQL version 10 is in PREVIEW*

The Open Service Broker for Azure contains the **Azure PostgreSQL Database** service shown:

| Service Type                  | Description                                                  |
| ----------------------------- | ------------------------------------------------------------ |
| `azure-postgresql-9.6-dbms`          | Provision both an Azure Database for PostgreSQL Database Management System (DBMS) and a database. |

The `azure-postgresql-9.6-dbms` provisions an Azure Database for PostgreSQL DBMS instance containing no databases. Databases can be created through subsequent provision requests using the `azure-postgresql-database` service.

>**NOTE:** This version of the service is based on [Open Service Broker for Azure](https://github.com/Azure/open-service-broker-azure).
For more details, read the **Plans Details** document.
