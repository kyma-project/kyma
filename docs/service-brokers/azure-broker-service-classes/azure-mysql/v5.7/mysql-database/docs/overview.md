# [Azure Database for MySQL](https://azure.microsoft.com/en-us/services/mysql/)

Open Service Broker for Azure contains three Azure Database for MySQL services. These services enable you to select the most appropriate provision scenario for your needs. These services are:

| Service Name | Description |
|--------------|-------------|
| `azure-mysql-5-7` | Provision both an Azure Database for MySQL Database Management System (DBMS) and a database, using MySQL 5.7 |
| `azure-mysql-5-7-dbms` | Provision only an Azure Database for MySQL DBMS with MySQL 5.7. This can be used to provision multiple databases at a later time. |
| `azure-mysql-5-7-database` | Provision a new database only upon a previously provisioned DBMS. |

The `azure-mysql-5-7` service allows you to provision both a DBMS and a database. When the provision operation is successful, the database will be ready to use. You can't provision additional databases onto an instance provisioned through this service. The `azure-mysql-5-7-dbms` and `azure-mysql-5-7-database` services, on the other hand, can be combined to provision multiple databases on a single DBMS.  For more information on each service, refer to the descriptions below.

>**NOTE:** This version of the service is based on Open Service Broker for Azure, version [1.3.1](https://github.com/Azure/open-service-broker-azure/releases).
For more details, read the **Plans Details** document.
