# Backing Up NFS Volumes in Google Cloud

> [!WARNING]
> This is a feature available only per request for SAP-internal teams.

This tutorial explains how to create backups for Network File System (NFS) volumes in Google Cloud.

## Prerequisites <!-- {docsify-ignore} -->

* You have the Cloud Manager module added.
* You have created a GcpNfsVolume. See [Use Network File System in Google Cloud](./01-20-20-gcp-nfs-volume.md).

> [!NOTE]
> All the examples below assume that the GcpNfsVolume is named `my-vol` and is in the same namespace as the GcpNfsVolumeBackup resource.

## Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create an GcpNfsVolumeBackup resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsVolumeBackup
   metadata:
     name: my-backup
   spec:
     source:
       volume:
         name: my-vol
   EOF
   ```

3. Wait for the GcpNfsVolumeBackup to be in the `Ready` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready gcpnfsvolumebackup/my-backup --timeout=300s
   ```

   Once the GcpNfsVolumeBackup is created, you should see the following message:

   ```console
   gcpnfsvolumebackup.cloud-resources.kyma-project.io/my-backup condition met
   ```

4. Observe the location of the created backup.

   ```shell
   kubectl -n $NAMESPACE get gcpnfsvolumebackup my-backup -o jsonpath='{.status.location}{"\n"}' 
   ```

## Next Steps

To clean up, remove the created GcpNfsVolumeBackup:

   ```shell
   kubectl delete -n $NAMESPACE gcpnfsvolumebackup my-backup
   ```
