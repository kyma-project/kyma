# Creating Scheduled Automatic NFS Volume Backups in Amazon Web Services

> [!WARNING]
> This is a feature available only per request for SAP-internal teams.

> [!WARNING]
> Long-running or frequent schedules can create too many backups and may result in cloud provider quota issues.
> For more information on how to avoid such issues, see [Scheduling Best Practices](../00-25-scheduling-best-practices.md).

This tutorial explains how to create scheduled automatic backups for Network File System (NFS) volumes in Amazon Web Services (AWS).

## Prerequisites <!-- {docsify-ignore} -->

* You have the Cloud Manager module added.
* You have created an AwsNfsVolume. See [Use Network File System in Amazon Web Services](./01-20-10-aws-nfs-volume.md).

> [!NOTE]
> All the examples below assume that the AwsNfsVolume is named `my-vol` and is in the same namespace as the AwsNfsBackupSchedule resource.

## Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create an AwsNfsBackupSchedule resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsNfsBackupSchedule
   metadata:
      name: my-backup-schedule
   spec:
      nfsVolumeRef:
        name: my-vol
      schedule: "0 * * * *"
      prefix: my-hourly-backup
      maxRetentionDays: 30
      maxReadyBackups: 100
      deleteCascade: true
   EOF
   ```

3. Wait for the AwsNfsVolumeBackup to be in the `Active` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=jsonpath='{.status.state}'=Active awsnfsbackupschedule/my-backup-schedule --timeout=300s
   ```

   Once the AwsNfsVolumeBackup is created, you should see the following message:

   ```console
   awsnfsbackupschedule.cloud-resources.kyma-project.io/my-backup-schedule condition met
   ```

4. Observe the nextRunTimes for creating the backups.

   ```shell
   kubectl -n $NAMESPACE get awsnfsbackupschedule my-backup-schedule -o jsonpath='{.status.nextRunTimes}{"\n"}' 
   ```

5. Wait till the time specified in the nextRunTimes (in the previous step) passes and see that the AwsNfsVolumeBackup objects get created.

   ```shell
   kubectl -n $NAMESPACE get awsnfsvolumebackup -l cloud-resources.kyma-project.io/scheduleName=my-backup-schedule 
   ```

## Next Steps <!-- {docsify-ignore} -->

To clean up, follow these steps:

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Remove the created schedule and the backups:
  
   ```shell
   kubectl delete -n $NAMESPACE awsnfsbackupschedule my-backup-schedule
   ```
