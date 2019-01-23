---
title: Services and Plans
type: Details
---

## Services & Plans

Service class contains the following plans and parameters:

| Plan Name | Description |
|-----------|-------------|
| `database` | New database on existing MySQL DBMS |

## Provision

Provisions a new database upon a previously provisioned DBMS. The new database will be named randomly.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **parentAlias** | `string` | Specifies the alias of the DBMS upon which the database should be provisioned. | Yes | |

## Bind

Creates a new user on the MySQL DBMS. The new user will be named randomly and
will be granted a wide array of permissions on the database.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| **host** | `string` | The fully-qualified address of the MySQL DBMS. |
| **port** | `int` | The port number to connect to on the MySQL DBMS. |
| **database** | `string` | The name of the database. |
| **username** | `string` | The name of the database user (in the form username@host). |
| **password** | `string` | The password for the database user. |
| **sslRequired** | `boolean` | Flag indicating if SSL is required to connect the MySQL DBMS. |
| **uri** | `string` | A URI string containing all necessary connection information. |
| **tags** | `string[]` | A list of tags consumers can use to identify the credential. |

## Unbind

Drops the applicable user from the MySQL DBMS.