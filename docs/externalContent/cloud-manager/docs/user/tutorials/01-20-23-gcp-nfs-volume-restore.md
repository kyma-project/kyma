# Restoring NFS Volume Backups in Google Cloud

> [!WARNING]
> This is a feature available only per request for SAP-internal teams.

This tutorial explains how to initiate a restore operation for the Network File System (NFS) volumes in Google Cloud. You can do it either using an existiong filestore or a new one.

## Prerequisites <!-- {docsify-ignore} -->

* You have the Cloud Manager module added.
* You have created a GcpNfsVolume. See [Use Network File System in Google Cloud](./01-20-20-gcp-nfs-volume.md).
* You have created a GcpNfsVolumeBackup. See [Back Up Network File System Volumes in Google Cloud](./01-20-21-gcp-nfs-volume-backup.md).

>[!NOTE]
>All the examples below assume that the GcpNfsVolume is named `my-vol`, the GcpNfsVolumeBackup is named `my-backup` 
and both are in the same namespace as the GcpNfsVolumeRestore resource.

## Use an Existing Filestore <!-- {docsify-ignore} -->

### Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable.

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create an GcpNfsVolumeRestore resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsVolumeRestore
   metadata:
     name: my-restore
   spec:
     source:
       backup:
         name: my-backup
         namespace: $NAMESPACE
     destination:
       volume:
         name: my-vol
         namespace: $NAMESPACE
   EOF
   ```

3. Wait for the GcpNfsVolumeRestore to be in the `Done` state and have the `Ready` condition.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready gcpnfsvolumerestore/my-restore --timeout=600s
   ```

   Once the GcpNfsVolumeRestore is completed, you should see the following message:

   ```console
   gcpnfsvolumerestore.cloud-resources.kyma-project.io/my-restore condition met
   ```

### Next Steps

To clean up, remove the created GcpNfsVolumeRestore:

   ```shell
   kubectl delete -n $NAMESPACE gcpnfsvolumerestore my-restore
   ```

## Create a New Filestore <!-- {docsify-ignore} -->

### Steps

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create a new GcpNfsVolume resource with `sourceBackup` referring to the existing backup.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsVolume
   metadata:
     name: my-vol2
   spec:
     location: us-west1-a
     capacityGb: 1024
     sourceBackup:
       name: my-backup
       namespace: $NAMESPACE
   EOF
   ```

3. Wait for the GcpNfsVolume to be in the `Ready` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready gcpnfsvolume/my-vol2 --timeout=600s
   ```

   Once the GcpNfsVolume is created, you should see the following message:

   ```console
   gcpnfsvolume.cloud-resources.kyma-project.io/my-vol2 condition met
   ```

### Next Steps

To clean up, remove the created GcpNfsVolume:

   ```shell
   kubectl delete -n $NAMESPACE gcpnfsvolume my-vol2
   ```
