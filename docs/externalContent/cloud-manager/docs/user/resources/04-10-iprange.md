# IpRange Custom Resource

The `iprange.cloud-resources.kyma-project.io` custom resource (CR) specifies the VPC network
IP range that is used for IP address allocation for cloud resources that require an IP address.

You are allowed to have one IpRange CR. If there are multiple IpRange resources in the cluster, the
oldest one is reconciled and the other is ignored and put into the `Error` state.

Once an IpRange CR is created and reconciled, the Cloud Manager controller reserves the specified IP range
in the Virtual Private Cloud (VPC) network of the cluster in the underlying cloud provider. The IP address from that range is
assigned to the provisioned resources of the cloud provider that require IP addresses. Once a 
cloud resource is assigned the local VPC network IP address it becomes functional and usable from the
cluster network and from the cluster workloads.

You don't have to create an IpRange resource. Once needed, it is automatically created
and Classless Inter-Domain Routing (CIDR) range automatically chosen adjacent to and with the same size as the cluster nodes IP range.
For most use cases this automatic allocation is sufficient.

You might be interested in manually creating an IpRange resource with specific CIDR in advanced cases of
VPC network topology when cluster and cloud resources are not the only resources in the network, so you
can avoid IP range collisions. 

IpRange can be deleted and deprovisioned only if there are no cloud resources using it. In other words,
an IpRange and its underlying VPC network address range can be purged only if there are no cloud resources
using an IP from that range.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

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

## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary IpRange custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: IpRange
metadata:
  name: my-range
spec:
  cidr: 10.250.4.0/22
```
