---
title: Services and Plans
type: Details
---

## Services & Plans

All of the `azure-mysql` service classes contains the following plans and parameters:

| Plan Name | Description |
|-----------|-------------|
| `Basic Tier` | Basic Tier, up to 2 vCores, variable I/O performance |
| `General Purpose Tier` | General Purpose Tier, up to 32 vCores, predictable I/O Performance, local or geo-redundant backups |
| `Memory Optimized Tier` | Memory Optimized Tier, up to 16 memory optimized vCores, predictable I/O Performance, local or geo-redundant backups |

### Provision

Provisions a new MySQL DBMS and a new database upon it. The new database will be named randomly.

#### Provisioning Parameters

The three plans each have additional provisioning parameters with different default and allowed values. See the tables below for details on each.

Provisioning Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | N | 1 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |

Provisioning Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16 or 32 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 2048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `backupRedundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |

Provisioning Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8 or 16 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 2048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `backupRedundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |

### Binding

Creates a new user on the MySQL DBMS. The new user will be named randomly and
will be granted a wide array of permissions on the database.

#### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| `host` | `string` | The fully-qualified address of the MySQL DBMS. |
| `port` | `int` | The port number to connect to on the MySQL DBMS. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user (in the form username@host). |
| `password` | `string` | The password for the database user. |
| `sslRequired` | `boolean` | Flag indicating if SSL is required to connect the MySQL DBMS. |
| `uri` | `string` | A URI string containing all necessary connection information. |
| `tags` | `string[]` | A list of tags consumers can use to identify the credential. |

### Examples

#### Kubernetes

The `contrib/k8s/examples/mysql/mysql-instance.yaml` can be used to provision the `basic` plan. This can be done with the following example:

```console
kubectl create -f contrib/k8s/examples/mysql/mysql-instance.yaml
```

You can then create a binding with the following command:

```console
kubectl create -f contrib/k8s/examples/mysql/mysql-binding.yaml
```
