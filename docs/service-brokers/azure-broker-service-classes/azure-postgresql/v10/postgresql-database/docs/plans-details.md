---
title: Services and Plans
type: Details
---

## Service description

The `azure-postgresql-10-database` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `database` | New database on existing DBMS |

## Provision

Provisions a new database upon an existing PostgreSQL DBMS. The new
database will be named randomly and will be owned by a role (group) of the same
name.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **extensions** | `string[]` | Specifies a list of PostgreSQL extensions to install | No | |
| **parentAlias** | `string` | Specifies the alias of the DBMS upon which the database should be provisioned. | Yes | |

**Note**: You should use corresponding  `dbms` service instance as the parent of `database` service instance.  For example, you should use `azure-postgresql-10-dbms` as the parent of `azure-postgresql-10-database`.

## Bind

Creates a new role (user) on the PostgreSQL DBNS. The new role will be named
randomly and added to the  role (group) that owns the database.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| **host** | `string` | The fully-qualified address of the PostgreSQL DBMS. |
| **port** | `int` | The port number to connect to on the PostgreSQL DBMS. |
| **database** | `string` | The name of the database. |
| **username** | `string` | The name of the database user (in the form username@host). |
| **password** | `string` | The password for the database user. |
| **sslRequired** | `boolean` | Flag indicating if SSL is required to connect the MySQL DBMS. |
| **uri** | `string` | A URI string containing all necessary connection information. |
| **tags` | `string[]` | A list of tags consumers can use to identify the credential. |
