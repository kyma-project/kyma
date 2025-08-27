# GcpNfsVolumeBackup Custom Resource

The `gcpnfsvolumebackup.cloud-resources.kyma-project.io` namespaced custom resource (CR) describes the GCP Filestore
instance's backup.
While the GCP Filestore backup is created in the underlying cloud provider subscription, it needs its source GCP 
Filestore instance to be available. But upon its creation, it can be used independently of the source instance.

GCP Filestore backups are regional resources, and they are created in the same region as the source GCP Filestore 
instance unless specified otherwise.

For a given Gcp Filestore, backups are incremental, as long as they are created on the same region. 
This reduces latency on backup creation. However, if a backup is created in a different region from the latest backup, 
it will be a full backup.
To learn more, read [Filestore Backup Creation](https://cloud.google.com/filestore/docs/backups#backup-creation).

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type                | Description                                                                                                                   |
|-----------------------------|---------------------|-------------------------------------------------------------------------------------------------------------------------------|
| **source**                  | object              | Required. Specifies the source of the backup.                                                                                 |
| **source.volume**           | object              | Required. Reference of the existing source GcpNfsVolume that is backed up.                                                    |
| **source.volume.name**      | string              | Required. Name of the source GcpNfsVolume.                                                                                    |
| **source.volume.namespace** | string              | Optional. Namespace of the source GcpNfsVolume. Defaults to the namespace of the GcpNfsVolumeBackup resource if not provided. |
| **location**                | string              | Optional. The Region where the backup resides. Defaults to the region of source GcpNfsVolume.                                 |

**Status:**

| Parameter                         | Type       | Description                                                                                                                          |
|-----------------------------------|------------|--------------------------------------------------------------------------------------------------------------------------------------|
| **state**                         | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
| **location**                      | string     | Signifies the location of the backup. This is particularly useful, if location is not provided in the spec.                          |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                                 |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                                |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                             |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                  |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                           |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                       |

## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary GcpNfsVolumeBackup custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackup
metadata:
  name: my-backup
spec:
  source:
    volume:
      name: my-vol
  location: us-west1
```
