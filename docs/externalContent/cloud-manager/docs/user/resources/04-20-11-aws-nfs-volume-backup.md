# AwsNfsVolumeBackup Custom Resource

The `awsnfsvolumebackup.cloud-resources.kyma-project.io` namespaced custom resource (CR) describes the AWS EFS Filesystem backup.
While the AWS EFS Filesystem backup is created in the underlying cloud provider subscription, it needs its source AWS 
EFS Filesystem instance to be available. But upon its creation, it can be used independently of the source instance.

For a given AWS EFS Filesystem, backups are incremental. This reduces latency on backup creation. 
To learn more, read [EFS Filesystem Backup Creation](https://docs.aws.amazon.com/efs/latest/ug/awsbackup.html).

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type                | Description                                                                                                                   |
|-----------------------------|---------------------|-------------------------------------------------------------------------------------------------------------------------------|
| **source**                  | object              | Required. Specifies the source of the backup.                                                                                 |
| **source.volume**           | object              | Required. Reference of the existing source AwsNfsVolume that is backed up.                                                    |
| **source.volume.name**      | string              | Required. Name of the source AwsNfsVolume.                                                                                    |
| **source.volume.namespace** | string              | Optional. Namespace of the source AwsNfsVolume. Defaults to the namespace of the AwsNfsVolumeBackup resource if not provided. |

**Status:**

| Parameter                         | Type       | Description                                                                                                                         |
|-----------------------------------|------------|-------------------------------------------------------------------------------------------------------------------------------------|
| **state**                         | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`.  |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                                |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                               |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                            |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                 |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                          |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                      |

## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary AwsNfsVolumeBackup custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolumeBackup
metadata:
  name: my-backup
spec:
  source:
    volume:
      name: my-vol
  location: us-west1
```
