# [Azure Database for PostgreSQL](https://azure.microsoft.com/en-us/services/postgresql/)

*Note: PostgreSQL version 9.6 is STABLE, and PostgreSQL version 10 is in PREVIEW*

Open Service Broker for Azure contains three types of Azure Database for PostgreSQL services. These services enable you to select the most appropriate provision scenario for your needs. These services are:

| Service Type                  | Description                                                  |
| ----------------------------- | ------------------------------------------------------------ |
| `azure-postgresql`          | Provision both an Azure Database for PostgreSQL Database Management System (DBMS) and a database. |
| `azure-postgresql-dbms`     | Provision only an Azure Database for PostgreSQL DBMS. This can be used to provision multiple databases at a later time. |
| `azure-postgresql-database` | Provision a new database only upon a previously provisioned DBMS. |

The `azure-postgresql-*` services allow you to provision both a DBMS and a database. When the provision operation is successful, the database will be ready to use. You can not provision additional databases onto an instance provisioned through these two services. The `azure-postgresql-*-dbms` and `azure-postgresql-*-database` services, on the other hand, can be combined to provision multiple databases on a single DBMS. Currently, OSBA supports two versions of Azure Database for PostgreSQL services:
<table>
	<thead>
		<tr>
			<th>Service Type</th>
			<th>Service name</th>
			<th>Stability</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td rowspan=2>azure-postgresql-*</td>
			<td>azure-postgresql-9-6</td>
			<td>Stable</td>
		</tr>
		<tr>
			<td>azure-postgresql-10</td>
			<td></td>
			<td>Preview</td>
		</tr>
		<tr>
			<td rowspan=2>azure-postgresql-*-dbms</td>
			<td>azure-postgresql-9-6-dbms</td>
			<td>Stable</td>
		</tr>
		<tr>
			<td>azure-postgresql-10-dbms</td>
			<td></td>
			<td>Preview</td>
		</tr>
		<tr>
			<td rowspan=2>azure-postgresql-*-database</td>
			<td>azure-postgresql-9-6-database</td>
			<td>Stable</td>
		</tr>
		<tr>
			<td>azure-postgresql-10-database</td>
			<td></td>
			<td>Preview</td>
		</tr>
	</tbody>
</table>

>**NOTE:** This version of the service is based on Open Service Broker for Azure, version [1.3.1](https://github.com/Azure/open-service-broker-azure/releases).
For more details, read the **Plans Details** document.
