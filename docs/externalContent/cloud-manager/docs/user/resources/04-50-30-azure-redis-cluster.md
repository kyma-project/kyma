# AzureRedisCluster Custom Resource

The `azurerediscluster.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the Azure Cache for Redis cluster.
Once the cluster is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as AzureRedisCluster.

The current implementation supports the Premium tier, which is explained in detail on the [Azure Cache for Redis overview page](https://azure.microsoft.com/en-us/products/cache).

> [!TIP] _Only for advanced cases of network topology_
> Redis requires 2 IP addresses per shard. IP addresses can be configured using the IpRange CR. Those IP addresses are allocated from the [IpRange CR](./04-10-iprange.md). If an IpRange CR is not specified in the AzureRedisCluster, then the default IpRange is used. If the default IpRange does not exist, it is automatically created. Manually create a non-default IpRange with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology when you want to control the network segments to avoid range conflicts with other networks.

When creating AzureRedisCluster, one field is mandatory: `redisTier`.

Optionally, you can specify the `redisConfiguration`, `redisVersion`, `shardCount`, `replicasPerPrimary` and `redisConfiguration` fields.

> [!NOTE]
> Non SSL port is disabled.

## Specification

This table lists the parameters of AzureRedisCluster, together with their descriptions:

| Parameter                                              | Type   | Description                                                                                                                                                                                                                                                                                                |
|--------------------------------------------------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ipRange**                                            | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created.                                                                                                                                                                           |
| **ipRange.name**                                       | string | Required. Name of the existing IpRange to use.                                                                                                                                                                                                                                                             | 
| **rediTier**                                           | string | Required. The service capacity of the cluster. Supported values are C3, C4, C5, C6 and C7.                                                                                                                                                                                                |
| **redisVersion**                                       | int    | Optional. The version of Redis software. Defaults to `6.0`.                                                                                                                                                                                                                                                |
| **shardCount**                                         | int    | Optional. The number of shards to be created on a Premium Cluster Cache.                                                                                                                                                                                                                                   |
| **replicasPerPrimary**                                 | int    | Optional. 	The number of replicas to be created per primary.                                                                                                                                                                                                                                               |
| **redisConfiguration**                                 | object | Optional. Object containing Redis configuration options.                                                                                                                                                                                                                                                   |
| **redisConfiguration.maxclients**                      | int    | Optional. Max number of Redis clients. Limited to [7,500 to 40,000.](https://azure.microsoft.com/en-us/pricing/details/cache/)                                                                                                                                                                             |
| **redisConfiguration.maxmemory-reserved**              | int    | Optional. [Configure your maxmemory-reserved setting to improve system responsiveness.](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-best-practices-memory-management#configure-your-maxmemory-reserved-setting)                                                                    |
| **redisConfiguration.maxmemory-delta**                 | int    | Optional. Gets or sets value in megabytes reserved for non-cache usage per shard e.g. failover.                                                                                                                                                                                                            | 
| **redisConfiguration.maxmemory-policy**                | int    | Optional. The setting for how Redis will select what to remove when **maxmemory** (the size of the cache offering you selected when you created the cache) is reached. Defaults to `volatile-lru`.                                                                                                         | 
| **redisConfiguration.maxfragmentationmemory-reserved** | int    | Optional. [Configure your maxmemory-reserved setting to improve system responsiveness.](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-best-practices-memory-management#configure-your-maxmemory-reserved-setting)                                                                    |
| **authSecret**                                         | object | Optional. Auth Secret options.                                                                                                                                                                                                                                                                             |
| **authSecret.name**                                    | string | Optional. Auth Secret name.                                                                                                                                                                                                                                                                                |
| **authSecret.labels**                                  | object | Optional. Auth Secret labels. Keys and values must be a string.                                                                                                                                                                                                                                            |
| **authSecret.annotations**                             | object | Optional. Auth Secret annotations. Keys and values must be a string.                                                                                                                                                                                                                                       |
| **authSecret.extraData**                               | object | Optional. Additional Secret Data entries. Keys and values must be a string. Allows users to define additional data fields that will be present in the Secret. The well-known data fields can be used as templates. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). |

## Auth Secret Details

The following table list the meaningful parameters of the auth Secret:

| Parameter                 | Type   | Description                                                                                     |
|---------------------------|--------|-------------------------------------------------------------------------------------------------|
| **.metadata.name**        | string | Name of the auth Secret. It shares the name with AzureRedisCluster unless specified otherwise. |
| **.metadata.labels**      | object | Specified custom labels (if any)                                                                |
| **.metadata.annotations** | object | Specified custom annotations (if any)                                                           |
| **.data.host**            | string | Primary connection host. Base64 encoded.                                                        |
| **.data.port**            | string | Primary connection port. Base64 encoded.                                                        |
| **.data.primaryEndpoint** | string | Primary connection endpoint. Provided in <host>:<port> format. Base64 encoded.                  |
| **.data.authString**      | string | Auth string. Base64 encoded.                                                                    |

## Notes
* Parameter `replicasPerPrimary` can not be changed after the cluster creation.
* The following table defines the value mapping between KYMA Redis cluster tier and Azure Redis cluster tier:

  | KYMA Redis cluster tier | Azure Redis cluster tier | Tier size (GB) |
  |-------------------------|--------------------------|----------------|
  | C3                      | P1                       | 6              |
  | C4                      | P2                       | 13             |
  | C5                      | P3                       | 26             |
  | C6                      | P4                       | 53             |
  | C7                      | P5                       | 160            |

## Sample Custom Resource

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRedisCluster
metadata:
  name: azureRedisClusterExample
spec:
  redisConfiguration:
    maxclients: "8"
  redisVersion: "6.0"
  redisTier: "C3"
  shardCount: "2"
```
