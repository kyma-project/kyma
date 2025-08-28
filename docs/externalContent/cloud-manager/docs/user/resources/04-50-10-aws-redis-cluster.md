# AwsRedisCluster Custom Resource

The `awsrediscluster.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the AWS ElastiCache Redis instance with cluster mode enabled.
After the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as the AwsRedisCluster, unless specified otherwise.

The AwsRedisCluster requires an IpRange CR. The size of IpRange is relative to the number of shards and replicas. Those IP addresses are allocated from the IpRange.
If the IpRange is not specified in the AwsRedisCluster, the default IpRange is used.
If a default IpRange does not exist, it is automatically created.
For more information, see [IpRange Custom Resource](./04-10-iprange.md).

When creating AwsRedisCluster, the `redisTier`, and `shardCount` fields are mandatory.

Optionally, you can specify the `replicasPerShard`, `engineVersion`, `authEnabled`, `parameters`, and `preferredMaintenanceWindow` fields.

As in-transit encryption is always enabled, communication with the Redis instance requires a trusted Certificate Authority (CA). You must install it on the container (e.g., using `apt-get install -y ca-certificates && update-ca-certificate`).

## Redis Cluster Tiers

| RedisTier | Capacity (GiB) | Network (up to Gbps) | Machine            |
| --------- | -------------- | -------------------- | ------------------ |
| C1        | 1.37           | 5                    | cache.t4g.small    |
| C2        | 3.09           | 5                    | cache.t4g.medium   |
| C3        | 6.38           | 12.5                 | cache.m7g.large    |
| C4        | 12.93          | 12.5                 | cache.m7g.xlarge   |
| C5        | 26.04          | 15                   | cache.m7g.2xlarge  |
| C6        | 52.26          | 15                   | cache.m7g.4xlarge  |
| C7        | 103.68         | 15                   | cache.m7g.8xlarge  |
| C8        | 209.55         | 30                   | cache.m7g.16xlarge |


## Specification

This table lists the parameters of AwsRedisCluster, together with their descriptions:

| Parameter                                         | Type   | Description                                                                                                                                                                                                 |
| --------------------------------------------------| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ipRange**                                       | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created.                                                                            |
| **ipRange.name**                                  | string | Required. Name of the existing IpRange to use.          |
| **redisTier**                                     | string | Required. The Redis tier of the instance. Supported values are `C1`, `C2`, `C3`, `C4`, `C5`, `C6`, `C7`, `C8`.        |
| **shardCount**                                    | int    | Required. Number of shards. Supported values are from `1` to `500`.     |
| **replicasPerShard**                              | int    | Optional. Number of replicas per shard. Supported values are from `0` to `5`. If left undefined, it defaults to `1`. Without replicas, a single shard failure can result in permanent data loss. |
| **engineVersion**                                 | string | Optional. Supported values are `"7.1"`, `"7.0"`, and `"6.x"`. Defaults to `"7.0"`. Can be upgraded. |
| **authEnabled**                                   | bool   | Optional. Enables using an AuthToken (password) when issuing Redis OSS commands. Defaults to `false`. |
| **parameters**                                    | object | Optional. Provided values are passed to the Redis configuration. Supported values can be read on [Amazons's Redis OSS-specific parameters page](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/ParameterGroups.Redis.html). If left empty, defaults to an empty object. |
| **preferredMaintenanceWindow**                    | string | Optional. Defines a desired window during which updates can be applied. If not provided, maintenance events can be performed at any time during the default time window. To learn more about maintenance window limitations and requirements, see [Managing maintenance](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/maintenance-window.html). |
| **authSecret**                                    | object | Optional. Auth Secret options.                                                                                                                                                                              |
| **authSecret.name**                               | string | Optional. Auth Secret name.                                                                                                                                                                                 |
| **authSecret.labels**                             | object | Optional. Auth Secret labels. Keys and values must be a string.                                                                                                                                             |
| **authSecret.annotations**                        | object | Optional. Auth Secret annotations. Keys and values must be a string.                                                                                                                                        |
| **authSecret.extraData**                          | object | Optional. Additional Secret Data entries. Keys and values must be a string. Allows users to define additional data fields that will be present in the Secret. The well-known data fields can be used as templates. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). |

## Auth Secret Details

The following table list the meaningful parameters of the auth Secret:

| Parameter                   | Type   | Description                                                                                                 |
| --------------------------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| **.metadata.name**          | string | Name of the auth Secret. It will share the name with the AwsRedisCluster unless specified otherwise        |
| **.metadata.labels**        | object | Specified custom labels (if any)                                                                            |
| **.metadata.annotations**   | object | Specified custom annotations (if any)                                                                       |
| **.data.host**              | string | Primary connection host.                                                                                    |
| **.data.port**              | string | Primary connection port.                                                                                    |
| **.data.primaryEndpoint**   | string | Primary connection endpoint. Provided in `<host>:<port>` format.                                              |
| **.data.authString**        | string | Auth string. Provided if authEnabled is set to true.                                                        |

## Sample Custom Resource

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsRedisCluster
metadata:
  name: awsrediscluster-sample
spec:
  redisTier: C1
  shardCount: 3
  replicasPerShard: 2
  engineVersion: "7.0"
  autoMinorVersionUpgrade: true
  authEnabled: true
  parameters:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
  preferredMaintenanceWindow: sun:23:00-mon:01:30
```
