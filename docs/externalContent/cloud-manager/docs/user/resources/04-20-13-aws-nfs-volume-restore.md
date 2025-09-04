# AwsNfsVolumeRestore Custom Resource

The `awsnfsvolumerestore.cloud-resources.kyma-project.io` namespaced custom resource (CR) describes the AWS EFS file 
system full restore operation on the same existing EFS file system.

Out-of-place restore and Item-Level restore are not supported by the `Cloud Manager`.

While the EFS file system restore operation is running in the underlying cloud provider subscription, it needs its 
source backup and its destination file system to be available. When the restore operation is finished, the restored file system is available on a recovery directory, named `aws-backup-restore_{datetime}`, off of the root directory.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                        | Type    | Description                                                                                                                                                                                                        |
|----------------------------------|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **source**                       | object  | Required. Specifies the source backup of the restore operation.                                                                                                                                                    |
| **source.backup**                | object  | Required. Reference of the existing AwsNfsVolumeBackup that is restored. The source volume of the AwsNfsVolumeBackup object that is referenced here, is used as the destination volume for the restore operation. |
| **source.backup.name**           | string  | Required. Name of the source AwsNfsVolumeBackup.                                                                                                                                                                   |
| **source.backup.namespace**      | string  | Optional. The namespace of the source AwsNfsVolumeBackup. Defaults to the namespace of the AwsNfsVolumeBackup resource if not provided.                                                                            |

**Status:**

| Parameter                         | Type       | Description                                                                                                                        |
|-----------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------|
| **state**                         | string     | Signifies the current state of **CustomObject**. Its value can be either `Done`, `Processing`, `Error`, `Failed`, or `InProgress`. |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                               |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                              |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                           |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                         |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                     |

## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary AwsNfsVolumeRestore custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolumeRestore
metadata:
  name: my-restore
  namespace: my-namespace
spec:
  source:
    backup:
      name: my-backup
```
