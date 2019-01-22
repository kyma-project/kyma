---
title: Services and Plans
type: Details
---

## Service description

The `azure-sql-12-0-dr-dbms-pair-registered` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `dbms` | Azure SQL Server DBMS-Only |

## Provision

Register a pair of SQL servers as a service instance: check the existence of these servers; check if the input administrator logins work. Databases with failover group can be created through subsequent provision requests using the `azure-sql-12-0-dr-database-pair` service.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `primaryResourceGroup` | `string` | The (new or existing) resource group with which to associate new resources. | Y | |
| `primaryLocation` | `string` | The Azure region in which to provision applicable resources. | Y | |
| `primaryServer` | `string` | The name of your existing primary server. | Y | |
| `primaryAdministratorLogin` | `string` | The administrator login of the primary server. | Y | |
| `primaryAdministratorLoginPassword` | `string` | The administrator login password of the primary server. | Y | |
| `secondaryResourceGroup` | `string` | The (new or existing) resource group with which to associate new resources. | Y | |
| `secondaryLocation` | `string` | The Azure region in which to provision applicable resources. | Y | |
| `secondaryServer` | `string` | The name of your existing secondary server. | Y | |
| `secondaryAdministratorLogin` | `string` | The administrator login of the secondary server. | Y | |
| `secondaryAdministratorLoginPassword` | `string` | The administrator login password of the secondary server. | Y | |
| `tags` | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | N | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |
| `alias` | `string` | Specifies an alias that can be used by later provision actions to create database pairs on this DBMS pair. | Y | |

## Update

Updates broker-stored administrator login/password in case you reset them.

### Updating Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `primaryAdministratorLogin` | `string` | The administrator login of the primary server. | N | |
| `primaryAdministratorLoginPassword` | `string` | The administrator login password of the primary server. | N | |
| `secondaryAdministratorLogin` | `string` | The administrator login of the secondary server. | N | |
| `secondaryAdministratorLoginPassword` | `string` | The administrator login password of the secondary server. | N | |

## Deprovision

Do nothing as it is a registered instance. If any database pairs have been provisioned on this DBMS pair, deprovisioning will be deferred until all databases have been deprovisioned.
