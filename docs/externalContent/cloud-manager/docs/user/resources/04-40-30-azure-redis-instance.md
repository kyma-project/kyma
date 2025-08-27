# AzureRedisInstance Custom Resource

The `azureRedisInstance.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the Azure Cache for Redis instance.
Once the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as AzureRedisInstance.

> [!TIP] _Only for advanced cases of network topology_
> Redis requires 2 IP addresses per shard. IP addresses can be configured using the IpRange CR. For more information, see [Configure a reserved IP address range](https://cloud.google.com/filestore/docs/creating-instances#configure_a_reserved_ip_address_range). Those IP addresses are allocated from the [IpRange CR](./04-10-iprange.md). If an IpRange CR is not specified in the AzureRedisInstance, then the default IpRange is used. If the default IpRange does not exist, it is automatically created. Manually create a non-default IpRange with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology when you want to control the network segments to avoid range conflicts with other networks.

When creating AzureRedisInstance, one field is mandatory: `redisTier`. 

In the **Kyma Standard** service tier, the instance does not have a replica. Thus, it cannot be considered highly available. 

| Kyma RedisTier | Capacity (GiB) | Azure RedisTier |
|----------------|----------------|-----------------|
| S1             | 1              | Basic C1        |
| S2             | 2.5            | Basic C2        |
| S3             | 6              | Basic C3        |
| S4             | 13             | Basic C4        |
| S5             | 26             | Basic C5        |

> [!NOTE]
> Kyma Standard S tier is mapped to Azure Basic C tier. It is cost-effective, and meant to be used for development purposes only.

> [!NOTE]
> [Basic tier is NOT recommended for production workloads](https://learn.microsoft.com/en-us/azure/well-architected/service-guides/azure-cache-redis/reliability#design-considerations), as it's not covered by [standard SLA](https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services).

In the **Kyma Premium** service tier, the instance comes with a read replica and automatic failover enabled. Thus, it can be considered highly available.

| Kyma RedisTier | Capacity (GiB) | Azure RedisTier |
|----------------|----------------|-----------------|
| P1             | 6              | Premium P1      |
| P2             | 13             | Premium P2      |
| P3             | 26             | Premium P3      |
| P4             | 53             | Premium P4      |
| P5             | 120            | Premium P5      |

Optionally, you can specify the `redisConfiguration`, `redisVersion`, and `redisConfiguration` fields.

> [!NOTE]
> Non SSL port is disabled.

## Specification

This table lists the parameters of AzureRedisInstance, together with their descriptions:

| Parameter                                              | Type   | Description                                                                                                                                                                                                                                                                                                 |
|--------------------------------------------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ipRange**                                            | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created.                                                                                                                                                                            |
| **ipRange.name**                                       | string | Required. Name of the existing IpRange to use.                                                                                                                                                                                                                                                              | 
| **rediTier**                                           | string | Required. The service capacity of the instance. Supported values are P1, P2, P3, P4, P5, S1, S2, S3, S4, S5.                                                                                                                                                                                                |
| **redisVersion**                                       | int    | Optional. The version of Redis software. Defaults to `6.0`.                                                                                                                                                                                                                                                 |
| **redisConfiguration**                                 | object | Optional. Object containing Redis configuration options.                                                                                                                                                                                                                                                    |
| **redisConfiguration.maxclients**                      | int    | Optional. Max number of Redis clients. Limited to [7,500 to 40,000.](https://azure.microsoft.com/en-us/pricing/details/cache/)                                                                                                                                                                              |
| **redisConfiguration.maxmemory-reserved**              | int    | Optional. [Configure your maxmemory-reserved setting to improve system responsiveness.](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-best-practices-memory-management#configure-your-maxmemory-reserved-setting)                                                                     |
| **redisConfiguration.maxmemory-delta**                 | int    | Optional. Gets or sets value in megabytes reserved for non-cache usage per shard e.g. failover.                                                                                                                                                                                                             | 
| **redisConfiguration.maxmemory-policy**                | int    | Optional. The setting for how Redis will select what to remove when **maxmemory** (the size of the cache offering you selected when you created the cache) is reached. Defaults to `volatile-lru`.                                                                                                          | 
| **redisConfiguration.maxfragmentationmemory-reserved** | int    | Optional. [Configure your maxmemory-reserved setting to improve system responsiveness.](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-best-practices-memory-management#configure-your-maxmemory-reserved-setting)                                                                     |
| **authSecret**                                    | object | Optional. Auth Secret options.                                                                                                                                                                                                                                                                              |
| **authSecret.name**                               | string | Optional. Auth Secret name.                                                                                                                                                                                                                                                                                 |
| **authSecret.labels**                             | object | Optional. Auth Secret labels. Keys and values must be a string.                                                                                                                                                                                                                                             |
| **authSecret.annotations**                        | object | Optional. Auth Secret annotations. Keys and values must be a string.                                                                                                                                                                                                                                        |
| **authSecret.extraData**                          | object | Optional. Additional Secret Data entries. Keys and values must be a string. Allows users to define additional data fields that will be present in the Secret. The well-known data fields can be used as templates. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). |

## Auth Secret Details

The following table list the meaningful parameters of the auth Secret:

| Parameter                 | Type   | Description                                                                                     |
|---------------------------|--------|-------------------------------------------------------------------------------------------------|
| **.metadata.name**        | string | Name of the auth Secret. It shares the name with AzureRedisInstance unless specified otherwise. |
| **.metadata.labels**      | object | Specified custom labels (if any)                                                                |
| **.metadata.annotations** | object | Specified custom annotations (if any)                                                           |
| **.data.host**            | string | Primary connection host. Base64 encoded.                                                        |
| **.data.port**            | string | Primary connection port. Base64 encoded.                                                        |
| **.data.primaryEndpoint** | string | Primary connection endpoint. Provided in <host>:<port> format. Base64 encoded.                  |
| **.data.authString**      | string | Auth string. Base64 encoded.                                                                    |

## Sample Custom Resource

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRedisInstance
metadata:
  name: azureRedisInstanceExample
spec:
  redisConfiguration:
    maxclients: "8"
  redisVersion: "6.0"
  redisTier: "P1"
```
