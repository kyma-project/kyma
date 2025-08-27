# Creating Scheduled Automatic NFS Volume Backups in Google Cloud

> [!WARNING]
> This is a feature available only per request for SAP-internal teams.

> [!WARNING]
> Long-running or frequent schedules can create too many backups and may result in cloud provider quota issues.
> For more information on how to avoid such issues, see [Scheduling Best Practices](../00-25-scheduling-best-practices.md).

This tutorial explains how to create scheduled automatic backups for Network File System (NFS) volumes in Google Cloud.

## Prerequisites <!-- {docsify-ignore} -->

* You have the Cloud Manager module added.
* You have created a GcpNfsVolume. See [Use Network File System in Google Cloud](./01-20-20-gcp-nfs-volume.md).

> [!NOTE]
> All the examples below assume that the GcpNfsVolume is named `my-vol` and is in the same namespace as the GcpNfsBackupSchedule resource.

## Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create a GcpNfsBackupSchedule resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsBackupSchedule
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

3. Wait for the GcpNfsVolumeBackup to be in the `Active` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=jsonpath='{.status.state}'=Active gcpnfsbackupschedule/my-backup-schedule --timeout=300s
   ```

   Once the GcpNfsVolumeBackup is created, you should see the following message:

   ```console
   gcpnfsbackupschedule.cloud-resources.kyma-project.io/my-backup-schedule condition met
   ```

4. Observe the nextRunTimes for creating the backups.

   ```shell
   kubectl -n $NAMESPACE get gcpnfsbackupschedule my-backup-schedule -o jsonpath='{.status.nextRunTimes}{"\n"}' 
   ```

5. Wait till the time specified in the nextRunTimes (in the previous step) and see that the GcpNfsVolumeBackup objects get created.

   ```shell
   kubectl -n $NAMESPACE get gcpnfsvolumebackup -l cloud-resources.kyma-project.io/scheduleName=my-backup-schedule 
   ```

## Next Steps

To clean up, remove the created schedule and the backups:

   ```shell
   kubectl delete -n $NAMESPACE gcpnfsbackupschedule my-backup-schedule
   ```
