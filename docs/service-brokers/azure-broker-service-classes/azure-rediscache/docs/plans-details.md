---
title: Services and Plans
type: Details
---

## Service description

The `azure-rediscache` service provides the following plan names and descriptions:

| Plan Name  | Description                      |
| ---------- | -------------------------------- |
| `basic`    | Basic Tier, default 250MB Cache  |
| `standard` | Standard Tier, default 1GB Cache |
| `premium`  | Premium Tier, default 6GB Cache  |

## Provision

This service provisions a new Redis cache.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name      | Type                | Description                                                  | Required | Default Value                                                |
| ------------------- | ------------------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| **location**          | `string`            | The Azure region in which to provision applicable resources. | Yes      |                                                              |
| **resourceGroup**     | `string`            | The (new or existing) resource group with which to associate new resources. | Yes      |                                                              |
| **skuCapacity**       | `integer`           | The size of the Redis cache to deploy.  Valid values: for C (Basic/Standard) family (0, 1, 2, 3, 4, 5, 6), for P (Premium) family (1, 2, 3, 4). They denotes real size  (250MB, 1GB, 2.5 GB, 6 GB, 13 GB, 26 GB, 53GB) and (6 GB, 13 GB, 26 GB, 53GB) respectively. | No       | If not provided, `0` is used for C (Basic/Standard) family; `1` is used for P (Premium) family. |
| **enableNonSslPort** | `string`            | Specifies whether the non-SSL Redis server port (6379) is enabled. Valid values: (`enabled`, `disabled`) | No       | If not provided, `disabled` is used. That is, you can't use non-SSL Redis server port by default. |
| **tags**              | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No       | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

For `premium` plan, following provisioning parameter is available:

| Parameter Name | Type      | Description                                                  | Required | Default Value                                          |
| -------------- | --------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------ |
| **shardCount**   | `integer` | The number of shards to be created on a Premium Cluster Cache. This action is irreversible. The number of shards can be changed later. | No       | If not specified, no additional shard will be created. |
| **subnetSettings**              | `object` | Setting to deploy the Redis cache inside a subnet, so that the  cache is only accessible in the subnet | No                                                   | If not specified, the Redis cache won't be deployed in a subnet, that is, the Redis cache is publicly addressable and the access is not limited to a particular VNet. |
| **subnetSettings.subnetId** | `string` | The full resource ID of a subnet in a virtual network to deploy the Redis cache in. The subnet should be in the same region with Redis cache. Example format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vn}/subnets/{sn} | Yes when `subnetSettings` is provided, otherwise no. |                                                              |
| **subnetSettings.staticIP**   | `string` | Static IP address. Required when deploying a Redis cache inside an existing Azure Virtual Network. Only valid when `subnetId` is provided. | No                                                   | If `staticIP` **is not** specified and `subnetId` **is** specified, one valid IP will be chosen randomly in the subnet. |
| **redisConfiguration**                                     | `object`  | Redis Settings. See below possible keys.                     | No                                                           | null object                                               |
| **redisConfiguration.rdb-backup-enabled**              | `string`  | Specifies whether RDB backup is enabled. Valid values: (`enabled`, `disabled`) | No                                                           | If not specified, RDB backup will be disabled by default. |
| **redisConfiguration.rdb-backup-frequency**            | `integer` | The frequency doing backup in minutes. Valid values: ( 15, 30, 60, 360, 720, 1440 ) | Yes when ` rdb-backup-enabled ` is set to `enabled`; otherwise is invalid. |                                                           |
| **redisConfiguration.rdb-storage-connection-string** | `string`  | The connnection string of the storage account for backup.    | Yes when ` rdb-backup-enabled ` is set to `enabled`; otherwise is invalid. |                                                           |

## Update 

Updates existing Redis cache.

### Updating parameters

| Parameter Name     | Type      | Description                                                  | Required |
| ------------------ | --------- | ------------------------------------------------------------ | -------- |
| **skuCapacity**      | `integer` | The size of the Redis cache to deploy.  Valid values: for C (Basic/Standard) family (0, 1, 2, 3, 4, 5, 6), for P (Premium) family (1, 2, 3, 4). They denotes real size  (250MB, 1GB, 2.5 GB, 6 GB, 13 GB, 26 GB, 53GB) and (6 GB, 13 GB, 26 GB, 53GB) respectively. **Note**: you can only update from a smaller capacity to a larger capacity, the reverse is not allowed. | No       |
| **enableNonSslPort** | `string`  | Specifies whether the non-ssl Redis server port (6379) is enabled. Valid values: (`enabled`, `disabled`) | No        |

For `premium` plan, following updating parameter is available:

| Parameter Name                                       | Type      | Description                                                  | Required                                                     |
| ---------------------------------------------------- | --------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| **shardCount**   | `integer` | The number of shards to be created on a Premium Cluster Cache. This action is irreversible. The number of shards can be changed later. **Note**: you can't update `skuCapacity` and `shardCount` at the same time. | N        |
| **redisConfiguration**                                 | `object`  | Redis Settings. See below possible keys.                     | No                                                           |
| **redisConfiguration.rdb-backup-enabled**            | `string`  | Specifies whether RDB backup is enabled. Valid values: (`enabled`, `disabled`) | No                                                           |
| **RedisConfiguration.rdb-backup-frequency**          | `integer` | The frequency doing backup in minutes. Valid values: ( 15, 30, 60, 360, 720, 1440 ) | Yes when `rdb-backup-enabled` is set to `enabled`; otherwise is invalid. |
| **redisConfiguration.rdb-storage-connection-string** | `string`  | The connnection string of the storage account for backup.    | Yes when `rdb-backup-enabled` is set to `enabled`; otherwise is invalid. |


### Credentials

Binding returns the following connection details and shared credentials:

| Field Name | Type     | Description                                       |
| ---------- | -------- | ------------------------------------------------- |
| **host**     | `string` | The fully-qualified address of the Redis cache.   |
| **port**     | `int`    | The port number to connect to on the Redis cache. |
| **password** | `string` | The password for the Redis cache.                 |
| **uri**      | `string` | The connection string for the Redis cache.        |

**Note**: if `enableNonSslPort` is set to `enabled`, then `port` will be `6379` and the scheme will be `redis` in `uri`; if `enableNonSslPort`  is set to `disabled`, then `port` will be `6380` and the scheme will be `rediss` in `uri`, and you can only use rediss to connect to the Redis cache.
