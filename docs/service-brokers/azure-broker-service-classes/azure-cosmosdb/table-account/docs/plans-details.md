---
title: Services and Plans
type: Details
---

## Service description

The `azure-cosmosdb-table-account` service consist of the following plan:

| Plan Name | Description |
|-----------|-------------|
| `account` | Database Account configured to use Table API |

## Provision

Provisions a new CosmosDB database account that can be accessed through any of the Azure Table API. The new database account is named using a new UUID.

### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `location` | `string` | The Azure region in which to provision applicable resources. | Y |  |
| `resourceGroup` | `string` | The (new or existing) resource group with which to associate new resources. | Y |  |
| `tags` | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | N | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |
| `ipFilters` | `object` | IP Range Filter to be applied to new CosmosDB account | N | A default filter is created that allows only Azure service access |
| `ipFilters.allowAccessFromAzure` | `string` | Specifies if Azure Services should be able to access the CosmosDB account. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | If left unspecified, defaults to enabled. |
| `ipFilters.allowAccessFromPortal` | `string` | Specifies if the Azure Portal should be able to access the CosmosDB account. If `allowAccessFromAzure` is set to enabled, this value is ignored. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | If left unspecified, defaults to enabled. |
| `ipFilters.allowedIPRanges` | `array` | Values to include in IP Filter. Can be IP Address or CIDR range. | N | If not specified, no additional values will be included in filters. |
| `consistencyPolicy` | `object` | The consistency policy for the Cosmos DB account. | N | |
| `consistencyPolicy.defaultConsistencyLevel` | `string` | The default consistency level and configuration settings of the Cosmos DB account. - Eventual, Session, BoundedStaleness, Strong, ConsistentPrefix | Y | |
| `consistencyPolicy.maxStalenessPrefix` | `integer` | When used with the Bounded Staleness consistency level, this value represents the number of stale requests tolerated. Accepted range for this value is 1 – 2,147,483,647. Required when defaultConsistencyPolicy is set to 'BoundedStaleness'. | N | |
| `consistencyPolicy.maxIntervalInSeconds` | `integer` | When used with the Bounded Staleness consistency level, this value represents the time amount of staleness (in seconds) tolerated. Accepted range for this value is 5 - 86400. Required when defaultConsistencyPolicy is set to 'BoundedStaleness'. | N | |
| `readRegions` | `array ` | Replication read regions to be created, your data will be synchronized across these regions, providing high availability and disaster recovery ability. Region's order in the array will be treated as failover priority.  See [here](#About Read Regions) for points to pay attention to. | N | If not specified, no replication region will be created. |
| `multipleWriteRegionsEnabled` | `string` | Specifies if you want  the account to write in multiple regions. Valid values are [ "enabled", "disabled"]. If set to "enabled", regions in `readRegions`  will also be writable. | N | If not specified, "disabled" will be used as the default value. |
| `autoFailoverEnabled` | `string ` | Specifies if you want Cosmos DB to perform automatic failover of the write region to one of the read regions in the rare event of a data center outage. Valid values are [ "enabled", "disabled"]. **Note**: If `multipleWriteRegionsEnabled` is set to `enabled`, all regions will be writable, and this attribute will not work. | N | If not specified, default "disabled". |

## Bind

Returns a copy of one shared set of credentials.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| `uri` | `string` | The fully-qualified address and port of the CosmosDB database account. |
| `primaryKey` | `string` | A secret key used for connecting to the CosmosDB database account. |
| `primaryConnectionString` | `string` | The full connection string, which includes the URI and primary key. |

## Deprovision

Deletes the CosmosDB database account.

## Update

Idempotently update the service instance to specified state.

| Parameter Name | Type                | Description                                                  | Required | Default Value                                                |
| -------------- | ------------------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| `tags` | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | N | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |
| `ipFilters` | `object` | IP Range Filter to be applied to new CosmosDB account | N | A default filter is created that allows only Azure service access |
| `ipFilters.allowAccessFromAzure` | `string` | Specifies if Azure Services should be able to access the CosmosDB account. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | If left unspecified, defaults to enabled. |
| `ipFilters.allowAccessFromPortal` | `string` | Specifies if the Azure Portal should be able to access the CosmosDB account. If `allowAccessFromAzure` is set to enabled, this value is ignored. Valid valued are `""` (unspecified), `enabled`, or `disabled`. | N | If left unspecified, defaults to enabled. |
| `ipFilters.allowedIPRanges` | `array` | Values to include in IP Filter. Can be IP Address or CIDR range. | N | If not specified, no additional values will be included in filters. |
| `readRegions` | `array ` | Read regions to be created, your data will be synchronized across these regions, providing high availability and disaster recovery ability. Region's order in the array will be treated as failover priority. **Note**: you can't update `readRegions` and other properties at the same time; if you want to update both `readRegions` and other properties, please update them in separate update operations. See [here](#About Read Regions) for points to pay attention to. | N | If not specified, no replication region will be created. |
| `autoFailoverEnabled` | `string ` | Specifies if you want Cosmos DB to perform automatic failover of the write region to one of the read regions in the rare event of a data center outage. Valid values are [ "enabled", "disabled"]. **Note**: If `multipleWriteRegionsEnabled` is set to `enabled`, all regions will be writable, and this attribute will not work. | N | If not specified, default "disabled". |
