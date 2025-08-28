# GcpRedisCluster Custom Resource

The `gcprediscluster.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).

It describes the [Google Memorystore Redis Cluster](https://cloud.google.com/memorystore/docs/cluster/memorystore-for-redis-cluster-overview) instance.

Once the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as the GcpRedisCluster, unless specified otherwise.

Redis Cluster requires a range of IP Addresses. They can be allocated using the GcpSubnet CR.
If a GcpSubnet CR is not specified in the GcpRedisCluster, then the default GcpSubnet is used.
If the default GcpSubnet does not exist, it is automatically created.
You can manually create a non-default GcpSubnet with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology when you want to control the network segments to avoid range conflicts with other networks.

When creating GcpRedisCluster, `redisTier`, and `shardCount` fields are mandatory.

Optionally, you can specify the `replicasPerShard` field.

As in-transit encryption is always enabled, communication with the Redis instance requires a certificate. The certificate can be found in the Secret on the `.data.CaCert.pem` path.


## Redis Tiers

In the **Standard** service tier, the instance does not have a replica. Thus, it cannot be considered highly available.

| RedisTier | Capacity (GiB) |
| --------- | -------------- |
| C1        | 1.4            |
| C3        | 6.5            |
| C4        | 13             |
| C6        | 58             |


## Specification

This table lists the parameters of GcpRedisCluster, together with their descriptions:

| Parameter                                         | Type   | Description                                                                                                                                                                                                 |
| --------------------------------------------------| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **subnet**                                        | object | Optional. GcpSubnet reference. If omitted, the default GcpSubnet is used. If the default GcpSubnet does not exist, it will be created.                                                                            |
| **subnet.name**                                   | string | Required. Name of the existing GcpSubnet to use.                                                                                                                                                              |
| **redisTier**                                     | string | Required. The Redis tier of the instance. Supported values are `C1`, `C3`, `C4`, `C6`.
| **shardCount**                                    | int    | Required. Number of shards. Minimum value is 1. Maximum number of shards is variable, and depends on selected number of `replciasPerShard`. Sum of all shards, and their respective replicas must be less or equal to 250.      |
| **replicasPerShard**                              | int    | Optional. Number of replicas per shard. Supported values are from `0` to `2`. If left undefined, it defaults to `1`. Without replicas, a single shard failure can result in permanent data loss.            |
| **authSecret**                                    | object | Optional. Auth Secret options.                                                                                                                                                                              |
| **authSecret.name**                               | string | Optional. Auth Secret name.                                                                                                                                                                                 |
| **authSecret.labels**                             | object | Optional. Auth Secret labels. Keys and values must be a string.                                                                                                                                             |
| **authSecret.annotations**                        | object | Optional. Auth Secret annotations. Keys and values must be a string.                                                                                                                                        |
| **authSecret.extraData**                          | object | Optional. Additional Secret Data entries. Keys and values must be a string. Allows users to define additional data fields that will be present in the Secret. The well-known data fields can be used as templates. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). |

## Auth Secret Details

The following table list the meaningful parameters of the auth Secret:

| Parameter                   | Type   | Description                                                                                                 |
| --------------------------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| **.metadata.name**          | string | Name of the auth Secret. It will share the name with the GcpRedisCluster unless specified otherwise        |
| **.metadata.labels**        | object | Specified custom labels (if any)                                                                            |
| **.metadata.annotations**   | object | Specified custom annotations (if any)                                                                       |
| **.data.host**              | string | Primary connection host.                                                                                    |
| **.data.port**              | string | Primary connection port.                                                                                    |
| **.data.primaryEndpoint**   | string | Primary connection endpoint. Provided in `<host>:<port>` format.                                              |
| **.data.authString**        | string | Auth string. Provided if authEnabled is set to true.                                                        |
| **.data.CaCert.pem**        | string | CA Certificate that must be used for TLS. Provided if transit encryption is set to server authentication.   |


## Sample Custom Resource

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpRedisCluster
metadata:
  name: gcprediscluster-sample
spec:
  redisTier: C1
  shardCount: 3
  replicasPerShard: 2
```
