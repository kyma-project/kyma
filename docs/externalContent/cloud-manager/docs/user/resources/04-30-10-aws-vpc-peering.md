# AwsVpcPeering Custom Resource

The `awsvpcpeering.cloud-resources.kyma-project.io` custom resource (CR) specifies the virtual network peering between Kyma and the remote AWS Virtual Private Cloud (VPC) network. Virtual network peering is only possible within the networks of the same cloud provider.

Once an `AwsVpcPeering` CR is created and reconciled, the Cloud Manager controller creates a VPC peering connection in the Kyma cluster underlying cloud provider landscape and accepts VPC peering connection in the remote cloud provider landscape.

## Specification

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                          | Type   | Description                                                                                                                      |
|------------------------------------|--------|----------------------------------------------------------------------------------------------------------------------------------|
| **remoteAccountId**                | string | Required. Specifies the Amazon Web Services account ID of the owner of the accepter VPC.                                         |
| **remoteRegion**                   | string | Required. Specifies the Region code for the accepter VPC.                                                                        |
| **remoteVpcId**                    | string | Required. Specifies the ID of the VPC with which you are creating the VPC peering connection                                     |
| **remoteRouteTableUpdateStrategy** | string | Optional. Specifies the remote route table update strategy. The value is one of the following: `AUTO`, `MATCHED`, `UNMATCHED`, or `NONE`. Defaults to `AUTO`. For more information, see [RemoteRouteTableUpdateStrategy](#RemoteRouteTableUpdateStrategy). <!-- markdown-link-check-disable-line --> |

**Status:**

| Parameter                         | Type       | Description                                                                                 |
|-----------------------------------|------------|---------------------------------------------------------------------------------------------|
| **id**                            | string     | Represents the VPC peering name on the Kyma cluster underlying cloud provider subscription. |
| **state**                         | string     | Signifies the current state of CustomObject.                                                |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                        |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                       |
| **conditions.message**            | string     | Provides more details about the condition status change.                                    |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                         |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.  |
| **conditions.type**               | string     | Provides a short description of the condition.                                              |

## RemoteRouteTableUpdateStrategy

To enable private Internet Protocol version 4 (IPv4) traffic between instances in peered VPC networks, Cloud Manager adds a peering route to the route tables associated with the remote VPC network. The route destination is the Classless Inter-Domain Routing (CIDR) block of the Kyma VPC network, and the target is the ID of the VPC peering connection.

Once VPC peering is established, Cloud Manager updates the route tables for the VPC peering connection. 

The `RemoteRouteTableUpdateStrategy` parameter specifies how Cloud Manager handles remote route tables:
- `AUTO` adds a peering route to all remote route tables.
- `MATCHED` adds a peering route to all remote route tables with the Kyma shoot name tag.
- `UNMATCHED` adds a peering route to all remote route tables without the Kyma shoot name tag.
- `NONE` does not interact with remote route tables.

## Sample Custom Resource

See an exemplary `AwsVpcPeering` custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsVpcPeering
metadata:
  name: peering-to-vpc-11122233
spec:
  remoteVpcId: vpc-11122233
  remoteRegion: us-west-2
  remoteAccountId: 123456789012
```
