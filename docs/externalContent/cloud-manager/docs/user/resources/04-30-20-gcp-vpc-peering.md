# GcpVpcPeering Custom Resource

The `gcpvpcpeering.cloud-resources.kyma-project.io` custom resource (CR) describes the Virtual Private Cloud (VPC) peering that you can create to allow communication between Kyma and a remote VPC in Google Cloud. It enables you to consume services available in the remote VPC from the Kyma cluster.

## Specification

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter              | Type   | Description                                                                                                                                                        |
|------------------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **importCustomRoutes** | bool   | If set to `true`, custom routes are exported from the remote VPC and imported into Kyma.                                                                           |
| **remotePeeringName**  | string | The VPC Peering name in the remote project. To find it, select **Google Cloud project under VPC > {VPC Name} > VPC Network Peering** in your Google Cloud Project. |
| **remoteProject**      | string | The Google Cloud project to be peered with Kyma. The remote VPC is located in this project.                                                                        |
| **remoteVpcName**      | string | The name of the remote VPC to be peered with Kyma.                                                                                                                 |

**Status:**

| Parameter                         | Type       | Description                                              |
|-----------------------------------|------------|----------------------------------------------------------|
| **state** (required)              | string     | Represents the current state of **CustomObject**.        |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.     |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.    |
| **conditions.message**            | string     | Provides more details about the condition status change. |
| **conditions.reason**             | string     | Defines the reason for the condition status change.      |
| **conditions.status** (required)  | string     | Represents the status of the condition.                  |
| **conditions.type**               | string     | Provides a short description of the condition.           |

## Sample Custom Resource

See an exemplary GcpVpcPeering custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpVpcPeering
metadata:
  name: "peering-with-kyma-dev"
spec:
  remotePeeringName: "peering-dev-vpc-to-kyma-dev"
  remoteProject: "my-remote-project"
  remoteVpc: "default"
  importCustomRoutes: false
```
