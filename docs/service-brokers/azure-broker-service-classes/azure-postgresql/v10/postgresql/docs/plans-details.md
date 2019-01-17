## Services & Plans

### Service: azure-postgresql-*

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, Up to 2 vCores, Variable I/O performance |
| `general-purpose` | General Purporse Tier, Up to 32 vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |
| `memory-optimized` | Memory Optimized Tier, Up to 16 memory optimized vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |

#### Behaviors

##### Provision

Provisions a new PostgreSQL DBMS and a new database upon that DBMS. The new
database will be named randomly and will be owned by a role (group) of the same
name.

###### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `location` | `string` | The Azure region in which to provision applicable resources. | Y | |
| `resourceGroup` | `string` | The (new or existing) resource group with which to associate new resources. | Y | |
| `sslEnforcement` | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | `""`. Left unspecified, SSL _will_ be enforced. |
| `firewallRules`  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | N | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| `firewallRules[n].name` | `string` | Specifies the name of the generated firewall rule |Y | |
| `firewallRules[n].startIPAddress` | `string` | Specifies the start of the IP range allowed by this firewall rule | Y | |
| `firewallRules[n].endIPAddress` | `string` | Specifies the end of the IP range allowed by this firewall rule | Y | |
| `tags` | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | N | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |
| `extensions` | `string[]` | Specifies a list of PostgreSQL extensions to install | N | |

####### Provisioning Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | N | 1 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |


####### Provisioning Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 32 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `backupRedundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |

####### Provisioning Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `backupRedundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |

##### Update

Updates a previously provisioned PostgreSQL DBMS. Currently updating the database extensions is not supported.

###### Updating Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `sslEnforcement` | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | `""`. Left unspecified, SSL _will_ be enforced. |
| `firewallRules`  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | N | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| `firewallRules[n].name` | `string` | Specifies the name of the generated firewall rule |Y | |
| `firewallRules[n].startIPAddress` | `string` | Specifies the start of the IP range allowed by this firewall rule | Y | |
| `firewallRules[n].endIPAddress` | `string` | Specifies the end of the IP range allowed by this firewall rule | Y | |

####### Updating Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | N | 1 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |


####### Updating Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 32 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |

####### Updating Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, or 16 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |

##### Bind

Creates a new role (user) on the PostgreSQL DBNS. The new role will be named
randomly and added to the  role (group) that owns the database.

###### Binding Parameters

This binding operation does not support any parameters.

###### Credentials

Binding returns the following connection details and credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| `host` | `string` | The fully-qualified address of the PostgreSQL DBMS. |
| `port` | `int` | The port number to connect to on the PostgreSQL DBMS. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user (in the form username@host). |
| `password` | `string` | The password for the database user. |
| `sslRequired` | `boolean` | Flag indicating if SSL is required to connect the MySQL DBMS. |
| `uri` | `string` | A URI string containing all necessary connection information. |
| `tags` | `string[]` | A list of tags consumers can use to identify the credential. |

##### Unbind

Drops the applicable role (user) from the PostgreSQL DBMS.

##### Deprovision

Deletes the PostgreSQL DBMS and database.

##### Examples

###### Kubernetes

The `contrib/k8s/examples/postgresql/postgresql-instance.yaml` can be used to provision the `basic50` plan. This can be done with the following example:

```console
kubectl create -f contrib/k8s/examples/postgresql/postgresql-instance.yaml
```

You can then create a binding with the following command:

```console
kubectl create -f contrib/k8s/examples/postgresql/postgresql-binding.yaml
```

### Service: azure-postgresql-*-dbms

| Plan Name | Description |
|-----------|-------------|
| `basic` | Basic Tier, Up to 2 vCores, Variable I/O performance |
| `general-purpose` | General Purporse Tier, Up to 32 vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |
| `memory-optimized` | Memory Optimized Tier, Up to 16 memory optimized vCores, Predictable I/O Performance, Local or Geo-Redundant Backups |

#### Behaviors

##### Provision

Provisions an Azure Database for PostgreSQL DBMS instance containing no databases. Databases can be created through subsequent provision requests using the `azure-postgresql-database` service.

###### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `location` | `string` | The Azure region in which to provision applicable resources. | Y | |
| `resourceGroup` | `string` | The (new or existing) resource group with which to associate new resources. | Y | |
| `alias` | `string` | Specifies an alias that can be used by later provision actions to create databases on this DBMS. | Y | |
| `sslEnforcement` | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | `""`. Left unspecified, SSL _will_ be enforced. |
| `firewallRules`  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | N | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| `firewallRules[n].name` | `string` | Specifies the name of the generated firewall rule |Y | |
| `firewallRules[n].startIPAddress` | `string` | Specifies the start of the IP range allowed by this firewall rule | Y | |
| `firewallRules[n].endIPAddress` | `string` | Specifies the end of the IP range allowed by this firewall rule | Y | |
| `tags` | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | N | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

####### Provisioning Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | N | 1 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |


####### Provisioning Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 32 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `backupRedundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |

####### Provisioning Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, or 16 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048 | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |
| `backupRedundancy` | `string` | Specifies the backup redundancy, either `local` or `geo` | N | `local` |

##### Update

Updates a previously provisioned PostgreSQL DBMS.

###### Updating Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `sslEnforcement` | `string` | Specifies whether the server requires the use of TLS when connecting. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | `""`. Left unspecified, SSL _will_ be enforced. |
| `firewallRules`  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | N | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| `firewallRules[n].name` | `string` | Specifies the name of the generated firewall rule |Y | |
| `firewallRules[n].startIPAddress` | `string` | Specifies the start of the IP range allowed by this firewall rule | Y | |
| `firewallRules[n].endIPAddress` | `string` | Specifies the end of the IP range allowed by this firewall rule | Y | |

####### Updating Parameters: basic

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 1 or 2 | N | 1 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |


####### Updating Parameters: general-purpose

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, 16, or 32 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |

####### Updating Parameters: memory-optimized

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `cores` | `integer` | Specifies vCores, which represent the logical CPU. Valid values are 2, 4, 8, or 16 | N | 2 |
| `storage` | `integer` | Specifies the amount of storage to allocate in GB. Ranges from 5 to 1048. Note, this must not be lower than what was given at provision time. | N | 10 |
| `backupRetention` | `integer` | Specifies the number of days to retain backups. Ranges from 7 to 35 | N | 7 |

##### Bind

This service is not bindable.

##### Unbind

This service is not bindable.

##### Deprovision

Deletes the PostgreSQL DBMS only. If databases have been provisioned on this DBMS, deprovisioning will be deferred until all databases have been deprovisioned.

##### Examples

###### Kubernetes

The `contrib/k8s/examples/postgresql/advanced/postgresql-dbms-instance.yaml` can be used to provision the `basic50` plan. This can be done with the following example:

```console
kubectl create -f contrib/k8s/examples/postgresql/advanced/postgresql-dbms-instance.yaml
```

### Service: azure-postgresql-*-database

| Plan Name | Description |
|-----------|-------------|
| `database` | New database on existing DBMS |

#### Behaviors

##### Provision

Provisions a new database upon an existing PostgreSQL DBMS. The new
database will be named randomly and will be owned by a role (group) of the same
name.

###### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `extensions` | `string[]` | Specifies a list of PostgreSQL extensions to install | N | |
| `parentAlias` | `string` | Specifies the alias of the DBMS upon which the database should be provisioned. | Y | |

**Note**: You should use corresponding  `dbms` service instance as the parent of `database` service instance.  For example, you should use `azure-postgresql-10-dbms` as the parent of `azure-postgresql-10-database`.

##### Update

Not currently supported.

##### Bind

Creates a new role (user) on the PostgreSQL DBNS. The new role will be named
randomly and added to the  role (group) that owns the database.

###### Binding Parameters

This binding operation does not support any parameters.

###### Credentials

Binding returns the following connection details and credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| `host` | `string` | The fully-qualified address of the PostgreSQL DBMS. |
| `port` | `int` | The port number to connect to on the PostgreSQL DBMS. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user (in the form username@host). |
| `password` | `string` | The password for the database user. |
| `sslRequired` | `boolean` | Flag indicating if SSL is required to connect the MySQL DBMS. |
| `uri` | `string` | A URI string containing all necessary connection information. |
| `tags` | `string[]` | A list of tags consumers can use to identify the credential. |

##### Examples

###### Kubernetes

The `contrib/k8s/examples/postgresql/postgresql-database-instance.yaml` can be used to provision the `database` plan. This can be done with the following example:

```console
kubectl create -f contrib/k8s/examples/postgresql/advanced/postgresql-database-instance.yaml
```

You can then create a binding with the following command:

```console
kubectl create -f contrib/k8s/examples/postgresql/advanced/postgresql-database-binding.yaml
```
