---
title: Services and Plans
type: Details
---

## Service description

The `azure-cosmosdb-sql-database` service consist of the following plan:

| Plan Name | Description |
|-----------|-------------|
| `database` | Database on existing CosmosDB database account configured to use SQL API |

## Provision

Provisions a new CosmosDB database onto an existing database account that can be accessed through any of the SQL API. The new database is named using a new UUID.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **parentAlias** | `string` | Specifies the alias of the CosmosDB database account upon which the database should be provisioned. | Yes | |

## Bind

Returns a copy of one shared set of credentials.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| **uri** | `string` | The fully-qualified address and port of the CosmosDB database account. |
| **primaryKey** | `string` | A secret key used for connecting to the CosmosDB database. |
| **primaryConnectionString** | `string` | The full connection string, which includes the URI and primary key. |
| **databaseName** | `string` | The generated database name. |
| **documentdb_database_id** | `string` | The database name provided in a legacy key for use with Azure libraries. |
| **documentdb_host_endpoint** | `string` | The fully-qualified address and port of the CosmosDB database account provided in a legacy key for use with Azure libraries. |
| **documentdb_master_key** | `string` | A secret key used for connecting to the CosmosDB database provided in a legacy key for use with Azure libraries. |

## Deprovision

Deletes the CosmosDB database. The existing database account is not deleted.
