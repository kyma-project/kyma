# Restoring NFS Volume Backups in Amazon Web Services

> [!WARNING]
> This is a feature available only per request for SAP-internal teams.

This tutorial explains how to initiate a restore operation for the ReadWriteMany (RWX) volumes in Amazon Web Services (AWS).

## Prerequisites <!-- {docsify-ignore} -->

* You have the Cloud Manager module added.
* You have created an AwsNfsVolume. See [Use Network File System in Amazon Web Services](./01-20-10-aws-nfs-volume.md).
* You have created an AwsNfsVolumeBackup. See [Back Up Network File System Volumes in Amazon Web Services](./01-20-11-aws-nfs-volume-backup.md).

>[!NOTE]
>The following examples assume that the AwsNfsVolumeBackup is named `my-backup` and is in the same namespace as the AwsNfsVolumeRestore resource.

## Steps <!-- {docsify-ignore} -->

### Restore on the Same or Existing Filestore <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable.

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create an AwsNfsVolumeRestore resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsNfsVolumeRestore
   metadata:
     name: my-restore
   spec:
     source:
       backup:
         name: my-backup
   EOF
   ```

3. Wait for the AwsNfsVolumeRestore to be in the `Done` state and have the `Ready` condition.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready awsnfsvolumerestore/my-restore --timeout=600s
   ```

   Once the AwsNfsVolumeRestore is completed, you should see the following message:

   ```console
   awsnfsvolumerestore.cloud-resources.kyma-project.io/my-restore condition met
   ```

## Next Steps

To clean up, remove the created AwsNfsVolumeRestore:

   ```shell
   kubectl delete -n $NAMESPACE awsnfsvolumerestore my-restore
   ```
