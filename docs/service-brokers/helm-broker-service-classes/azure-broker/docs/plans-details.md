---
title: Services and Plans
type: Details
---

## Service description

The `Azure Service Broker` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Default` | Installs the Azure Service Broker in a default configuration. |

## Provisioning

Provisioning an instance creates a new Azure Service Broker in the given Namespace and registers it in the Service Catalog.

### Provisioning parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **minimumStability** | `enum` | Minimum stability required for a module's services and plans to be listed in the broker's catalog. | YES | `Preview` |

These are the minimum stability values:
 * `Stable` - represents relative stability of the mature, production ready service modules:
    * [Azure Database for MySQL](https://github.com/Azure/open-service-broker-azure/blob/v1.4.0/docs/modules/mysql.md)
    * [Azure Database for PostgreSQL v9.6](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/postgresql.md)
    * [Azure SQL Database](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/mssql.md)
 * `Preview` - represents relative stability of modules that are approaching a stable state:
     * [Azure CosmosDB](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/cosmosdb.md)
     * [Azure Redis Cache](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/rediscache.md)
     * [Azure Database for PostgreSQL v10](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/postgresql.md)
     * [Azure Storage](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/storage.md)
 * `Experimental` - represents relative stability of the most immature service modules: 
    * [Azure Application Insights](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/appinsights.md)
    * [Azure Event Hubs](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/eventhubs.md)
    * [Azure IoT Hub](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/iothub.md)
    * [Azure Key Vault](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/keyvault.md)
    * [Azure Service Bus](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/servicebus.md)
    * [Azure Text Analytics (Cognitive Services)](https://github.com/Azure/open-service-broker-azure/tree/v1.4.0/docs/modules/textanalytics.md)

## Binding

Binding to this Service Class is disabled.

## Deprovisioning

Deprovisioning uninstalls the Azure Service Broker and unregisters it from the Service Catalog.
