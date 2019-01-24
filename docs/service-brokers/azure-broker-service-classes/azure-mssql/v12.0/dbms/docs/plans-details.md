---
title: Services and Plans
type: Details
---

## Service description

The `azure-sql-12-0-dbms` service provides the following plan:

| Plan Name | Description |
|-----------|-------------|
| `dbms` | Azure SQL Server DBMS-Only |

## Provision

Provisions a SQL Server DBMS instance containing no databases. Databases can be created through subsequent provision requests using the `azure-sql-12-0-database` service.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **location** | `string` | The Azure region in which to provision applicable resources. | Yes | |
| **resourceGroup** | `string` | The (new or existing) resource group with which to associate new resources. | Yes | |
| **tags** | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |
| **alias** | `string` | Specifies an alias that can be used by later provision actions to create databases on this DBMS. | Yes | |
| **firewallRules**  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | No | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| **firewallRules[n].name** | `string` | Specifies the name of the generated firewall rule |Y | |
| **firewallRules[n].startIPAddress** | `string` | Specifies the start of the IP range allowed by this firewall rule | Yes | |
| **firewallRules[n].endIPAddress** | `string` | Specifies the end of the IP range allowed by this firewall rule | Yes | |
| **connectionPolicy** | `string` | Changes connection policy if you want. Refer to [here](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-connectivity-architecture#connection-policy). Valid values are "Redirect", "Proxy", and "Default". | No | |

## Update

Updates a previously provisioned SQL DB DBMS.

### Updating Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **firewallRules**  | `array` | Specifies the firewall rules to apply to the server. Definition follows. | No | `[]` Left unspecified, Firewall will default to only Azure IPs. If rules are provided, they must have valid values. |
| **firewallRules[n].name** | `string` | Specifies the name of the generated firewall rule |Y | |
| **firewallRules[n].startIPAddress** | `string` | Specifies the start of the IP range allowed by this firewall rule | Yes | |
| **firewallRules[n].endIPAddress** | `string` | Specifies the end of the IP range allowed by this firewall rule | Yes | |
| **connectionPolicy** | `string` | Changes connection policy if you want. Refer to [here](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-connectivity-architecture#connection-policy). Valid values are "Redirect", "Proxy", and "Default". | No | |

## Deprovision

Deprovision will delete the SQL Server DBMS. If any databases have been provisioned on this DBMS, deprovisioning will be deferred until all databases have been deprovisioned.
