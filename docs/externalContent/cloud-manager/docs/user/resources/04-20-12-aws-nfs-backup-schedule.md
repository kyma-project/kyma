# AwsNfsBackupSchedule Custom Resource

The `awsnfsbackupschedule.cloud-resources.kyma-project.io` custom resource (CR) represents the user-defined schedule for creating a backup of the `AwsNfsVolume` instances at regular intervals. The CR performs the following actions:

- Creates the backups by creating the `awsnfsvolumebackup.cloud-resources.kyma-project.io` resources at the specified interval.
- Enables you to specify days and times in the form of CRON expressions to automatically create the backups.
- Automatically deletes the backups when the backup reaches the configured maximum retention days value.
- Enables you to temporarily suspend or resume the backup creation/deletion.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type                | Description                                                                                                                                                                                                                                                                           |
|-----------------------------|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **nfsVolumeRef**            | object              | Required. AwsNfsVolume reference.                                                                                                                                                                                                                                                     |
| **nfsVolumeRef.name**       | string              | Required. Name of the existing AwsNfsVolume.                                                                                                                                                                                                                                          |
| **nfsVolumeRef.namespace**  | string              | Optional. The namespace of the existing AwsNfsVolume.  Defaults to the namespace of the AwsNfsBackupSchedule resource if not provided.                                                                                                                                                |
| **schedule**                | string              | Optional. CRON type expression for the schedule. When this value is empty or not specified, this schedule runs only once at the specified start time. See also [Schedule Syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#schedule-syntax).               |
| **prefix**                  | string              | Optional. Prefix for the name of the created `AwsNfsVolumeBackup` resources. Defaults to name of this schedule.                                                                                                                                                                       |
| **startTime**               | metav1.Time         | Optional. Start time for the schedule. Value cannot be from the past. When not specified, the schedule becomes effective immediately.                                                                                                                                                 |
| **endTime**                 | metav1.Time         | Optional. End time for the schedule. Value cannot be from the past or before the `startTime`. When not specified, the schedule runs indefinitely.                                                                                                                                     |
| **maxRetentionDays**       | int                 | Optional. Maximum number of days to retain the backup resources. If not specified, the default value is 375 days. If `deleteCascade` is `true` for this schedule, then all the backups are deleted when the schedule is deleted irrespective of this configuration value. |
| **maxReadyBackups**        | int                 | Optional. Maximum number of backups in `Ready` state to be retained. Default value is 100.                                                                                                                                                                                |
| **maxFailedBackups**       | int                 | Optional. Maximum number of backups in `Failed` state to be retained. Default value is 5.                                                                                                                                                                                 |
| **suspend**                 | boolean             | Optional. Specifies whether or not to suspend the schedule temporarily. Defaults to `false`.                                                                                                                                                                                          |
| **deleteCascade**           | boolean             | Optional. Specifies whether to cascade delete the backup resources when this schedule is deleted. Defaults to `false`.                                                                                                                                                                |

**Status:**

| Parameter                         | Type                | Description                                                                                                                                                |
|-----------------------------------|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required)              | string              | Signifies the current state of **CustomObject**. Contains one of the following states:  `Processing`, `Pending`, `Suspended`, `Active`, `Done` or `Error`. |
| **conditions**                    | \[\]object          | Represents the current state of the CR's conditions.                                                                                                       |
| **conditions.lastTransitionTime** | string              | Defines the date of the last condition status change.                                                                                                      |
| **conditions.message**            | string              | Provides more details about the condition status change.                                                                                                   |
| **conditions.reason**             | string              | Defines the reason for the condition status change.                                                                                                        |
| **conditions.status** (required)  | string              | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                                                 |
| **conditions.type**               | string              | Provides a short description of the condition.                                                                                                             |
| **nextRunTimes**                  | \[\]string          | Provides the preview of the times when the next backups are created.                                                                                   |
| **nextDeleteTimes**               | map\[string\]string | Provides the backup objects and their expected deletion time (calculated based on `maxRetentionDays`).                                                     |
| **lastCreateRun**                 | string              | Provides the time when the last backup was created.                                                                                                        |
| **lastCreatedBackup**             | objectRef           | Provides the object reference of the last created backup.                                                                                                  |
| **lastDeleteRun**                 | string              | Provides the time when the last backup was deleted.                                                                                                        |
| **lastDeletedBackups**            | \[\]objectRef       | Provides the object references of the last deleted backups.                                                                                                |
| **schedule**                      | string              | Provides the CRON expression of the current active schedule.                                                                                               |
| **backupIndex**                   | int                 | Provides the current index of the backup created by this schedule.                                                                                         |
| **backupCount**                   | int                 | Provides the the number of backups currently present in the system.                                                                                        |

## Sample Custom Resource <!-- {docsify-ignore} -->

See an example of the AwsNfsBackupSchedule CR:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsBackupSchedule
metadata:
  name: aws-nfs-backup-schedule
  namespace: kyma-dev
spec:
  nfsVolumeRef:
    name: aws-nfs-sample-01
    namespace: kyma-dev
  schedule: "0 0 * * *"
  prefix: aws-nfs-daily-backup
  startTime: 2024-11-01T00:00:00Z
  endTime: 2025-12-31T00:00:00Z
  maxRetentionDays: 365
  maxReadyBackups: 150
  suspend: false
  deleteCascade: true
```
