# AwsRedisInstance Custom Resource

The `awsredisinstance.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the AWS ElastiCache Redis instance.
After the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as the AwsRedisInstance, unless specified otherwise.

The current implementation creates a single node replication group with cluster mode disabled.

The AwsRedisInstance requires an `/28` IpRange. Those IP addresses are allocated from the [IpRange](./04-10-iprange.md).
If the IpRange is not specified in the AwsRedisInstance, the default IpRange is used.
If a default IpRange does not exist, it is automatically created.
Manually create a non-default IpRange with specified CIDR and use it only in advanced cases of network topology when you want to be in control of the network segments to avoid range conflicts with other networks.

When creating AwsRedisInstance, only the `redisTier` field is mandatory.
It specifies the service tier (**Standard** or **Premium**), and the capacity tier.
Read on for more details.

Optionally, you can specify the `engineVersion`, `authEnabled`, `parameters`, and `preferredMaintenanceWindow` fields.

As in-transit encryption is always enabled, communication with the Redis instance requires a trusted Certificate Authority (CA). You must install it on the container (e.g., using `apt-get install -y ca-certificates && update-ca-certificate`).

## Redis Tiers

### Standard Tier

In the **Standard** service tier, the instance does not have a replica. Thus, it cannot be considered highly available.
The table below showcases which AWS machines are used for each tier.

| RedisTier | Capacity (GiB) | Network (up to Gbps) | Machine            |
| --------- | -------------- | -------------------- | ------------------ |
| S1        | 1.37           | 5                    | cache.t4g.small    |
| S2        | 3.09           | 5                    | cache.t4g.medium   |
| S3        | 6.38           | 12.5                 | cache.m7g.large    |
| S4        | 12.93          | 12.5                 | cache.m7g.xlarge   |
| S5        | 26.04          | 15                   | cache.m7g.2xlarge  |
| S6        | 52.26          | 15                   | cache.m7g.4xlarge  |
| S7        | 103.68         | 15                   | cache.m7g.8xlarge  |
| S8        | 209.55         | 30                   | cache.m7g.16xlarge |

### Premium Tier

In the **Premium** service tier, the instance comes with a read replica and automatic failover enabled. Thus, it can be considered highly available.
The table below showcases which AWS machines are used for each tier.

| RedisTier | Capacity (GiB) | Network (up to Gbps) | Machine            |
| --------- | -------------- | -------------------- | ------------------ |
| P1        | 6.38           | 12.5                 | cache.m7g.large    |
| P2        | 12.93          | 12.5                 | cache.m7g.xlarge   |
| P3        | 26.04          | 15                   | cache.m7g.2xlarge  |
| P4        | 52.26          | 15                   | cache.m7g.4xlarge  |
| P5        | 103.68         | 15                   | cache.m7g.8xlarge  |
| P6        | 209.55         | 30                   | cache.m7g.16xlarge |

## Specification

This table lists the parameters of AwsRedisInstance, together with their descriptions:

| Parameter                                         | Type   | Description                                                                                                                                                                                                 |
| --------------------------------------------------| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ipRange**                                       | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created.                                                                            |
| **ipRange.name**                                  | string | Required. Name of the existing IpRange to use.                                                                                                                                                              |
| **redisTier**                                     | string | Required. The Redis tier of the instance. Supported values are `S1`, `S2`, `S3`, `S4`, `S5`, `S6`, `S7`, `S8` for the **Standard** offering, and `P1`, `P2`, `P3`, `P4`, `P5`, `P6` for the **Premium** offering. |
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
| **.metadata.name**          | string | Name of the auth Secret. It will share the name with the AwsRedisInstance unless specified otherwise        |
| **.metadata.labels**        | object | Specified custom labels (if any)                                                                            |
| **.metadata.annotations**   | object | Specified custom annotations (if any)                                                                       |
| **.data.host**              | string | Primary connection host.                                                                                    |
| **.data.port**              | string | Primary connection port.                                                                                    |
| **.data.primaryEndpoint**   | string | Primary connection endpoint. Provided in <host>:<port> format.                                              |
| **.data.authString**        | string | Auth string. Provided if authEnabled is set to true.                                                        |

## Sample Custom Resource

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsRedisInstance
metadata:
  name: awsredisinstance-sample
spec:
  redisTier: P1
  engineVersion: "7.0"
  autoMinorVersionUpgrade: true
  authEnabled: true
  parameters:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
  preferredMaintenanceWindow: sun:23:00-mon:01:30
```
