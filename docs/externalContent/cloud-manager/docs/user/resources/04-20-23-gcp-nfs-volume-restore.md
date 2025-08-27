# GcpNfsVolumeRestore Custom Resource

The `gcpnfsvolumerestore.cloud-resources.kyma-project.io` namespaced custom resource (CR) describes the GCP Filestore
instance's restore operation on the same or an existing Filestore. This operation is only supported for BASIC tiers.
To learn more, read [Supported tiers](https://cloud.google.com/filestore/docs/backup-restore).

To restore a backup of a ZONAL or REGIONAL Filestore, the restore operation must be performed while a new Filestore instance is created.
This is supported by the `sourceBackup` field in the spec of `gcpnfsvolume.cloud-resources.kyma-project.io` CRD. To learn more, read [GcpNfsVolume Custom Resource](./04-20-20-gcp-nfs-volume.md).

While the GCP Filestore restore operation is running in the underlying cloud provider subscription, it needs its source GCP 
Filestore backup and its destination GCP Filestore instance to be available. Upon its completion, the GCP Filestore instance
is restored to the state of the source GCP Filestore backup.

Restore on the same Filestore means that the source Filestore of the backup is the same as the destination Filestore of the restore operation.
Restore on an existing Filestore means that the source Filestore of the backup is different from the destination Filestore of the restore operation 
and the destination Filestore of the restore operation already exists.

The capacity of the destination Filestore instance must be equal to or greater than the capacity of the source Filestore of the backup.

To learn more, read [Filestore Backup/Restore limitations](https://cloud.google.com/filestore/docs/backups#limitations-storage).

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                        | Type    | Description                                                                                                                             |
|----------------------------------|---------|-----------------------------------------------------------------------------------------------------------------------------------------|
| **source**                       | object  | Required. Specifies the source backup of the restore operation.                                                                         |
| **source.backup**                | object  | Required. Reference of the existing GcpNfsVolumeBackup that is restored.                                                                |
| **source.backup.name**           | string  | Required. Name of the source GcpNfsVolumeBackup.                                                                                        |
| **source.backup.namespace**      | string  | Required. The namespace of the source GcpNfsVolumeBackup. Defaults to the namespace of the GcpNfsVolumeBackup resource if not provided. |
| **destination**                  | object  | Required. Specifies the destination of the restore operation.                                                                           |
| **destination.volume**           | object  | Required. Reference of the existing GcpNfsVolume that is restored.                                                                      |
| **destination.volume.name**      | string  | Required. Name of the destination GcpNfsVolume.                                                                                         |
| **destination.volume.namespace** | string  | Optional. The namespace of the destination GcpNfsVolume. Defaults to the namespace of the GcpNfsVolumeRestore resource if not provided. |

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

See an exemplary GcpNfsVolumeRestore custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackup
metadata:
  name: my-restore
spec:
  source:
    backup:
      name: my-backup
      namespace: my-namespace
  destination:
    volume:
      name: my-vol
      namespace: my-namespace
```
