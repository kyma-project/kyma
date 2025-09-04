# GcpSubnet Custom Resource

The `gcpsubnet.cloud-resources.kyma-project.io` is a cluster-scoped custom resource (CR) that specifies the VPC Network Subnet. This resource is only available when the cluster cloud provider is Google Cloud Platform.
Currently, its only use is IP address allocation for the `GcpRedisCluster` CR.

Once a GcpSubnet CR is created and reconciled, the Cloud Manager controller creates a Subnet with defined CIDR
in the Virtual Private Cloud (VPC) Network of the cluster.

You don't have to create a GcpSubnet resource. Once needed, it is automatically created with the hardcoded CIDR `10.251.0.0/22`. For most use cases, this automatic allocation is sufficient.

You can manually create a GcpSubnet resource with a specific CIDR in advanced cases of VPC network topology. To avoid IP range collisions, create GcpSubnet when cluster and cloud resources are not the only resources in the network.

GcpSubnet can be deleted and deprovisioned only if no cloud resources are using it. In other words,
a GcpSubnet and its underlying VPC Network Subnet address range can be purged only if no cloud resources are using an IP from that range.

## Specification

This table lists the parameters of the GcpSubnet resource together with their descriptions:

**Spec:**

| Parameter | Type   | Description                                                                          |
|-----------|--------|--------------------------------------------------------------------------------------|
| **cidr**  | string | Specifies the CIDR of the IP range that will be allocated. For example, 10.250.4.0/22. |

**Status:**

| Parameter                         | Type       | Description                                                                                                                        |
|-----------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required)              | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                               |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                              |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                           |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                         |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                     |

## Sample Custom Resource

See an exemplary GcpSubnet custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpSubnet
metadata:
  name: my-gcpsubnet
spec:
  cidr: 10.252.8.0/22
```
