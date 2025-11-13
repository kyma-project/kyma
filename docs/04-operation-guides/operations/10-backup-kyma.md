# Back Up Kyma

## Context

The Kyma cluster load consists of Kubernetes [objects](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/) and [volumes](https://kubernetes.io/docs/concepts/storage/volumes/).

### Object Backup

Kyma relies on a managed Kubernetes cluster for periodic backups of Kubernetes objects to avoid any manual steps.

> [!WARNING]
> Automatic backup doesn't include Kubernetes volumes. Back up your volumes periodically either on demand, or set up a periodic job.

For example, Gardener uses [etcd](https://etcd.io/) as the Kubernetes backing store for all cluster data. Gardener runs periodic jobs to take major and minor snapshots of the etcd database to include Kubernetes objects in the backup.

The major snapshot that includes all resources is taken on a daily basis, while minor snapshots are taken every five minutes.

If the etcd database experiences any problems, Gardener automatically restores the Kubernetes cluster using the most recent snapshot.

### Volume Backup

We recommend that you back up your volumes periodically with the [VolumeSnapshot API resource](https://kubernetes.io/docs/concepts/storage/volume-snapshots/#volumesnapshots), which is provided by Kubernetes. You can use your snapshot to provision a new volume prepopulated with the snapshot data, or restore the existing volume to the state represented by the snapshot.

Taking volume snapshots is possible thanks to [Container Storage Interface (CSI) drivers](https://kubernetes-csi.github.io/docs/), which allow third-party storage providers to expose storage systems in Kubernetes. The driver must be specified in the VolumeSnapshotClass resource. Kyma clusters usually have a default VolumeSnapshotClass available. If you use the default resource, you don't have to configure the driver.

## Create On-Demand Volume Snapshots

To manually back up your volumes, use the [VolumeSnapshot](https://kubernetes.io/docs/concepts/storage/volume-snapshots/) Kubernetes resource:

1. Create a VolumeSnapshot resource using the default VolumeSnapshotClass and your PVC name:

   ```yaml
   kubectl apply -n {NAMESPACE} -f <<EOF
   apiVersion: snapshot.storage.k8s.io/v1
   kind: VolumeSnapshot
   metadata:
     name: snapshot
   spec:
     volumeSnapshotClassName: default
     source:
       persistentVolumeClaimName: {YOUR_PVC_NAME}
   EOF
   ```

   The VolumeSnapshot resource is created.

2. To verify that the snapshot was taken successfully, run `kubectl get -n {NAMESPACE} volumesnapshot -w` and check that the field **READYTOUSE** has status `true`.

## Back Up Resources Using Third-Party Tools

> [!WARNING]
> Third-party tools like Velero are not currently supported. These tools may have limitations and might not fully support automated cluster backups. They often require specific access rights to cluster infrastructure, which may not be available in Kyma's managed offerings, where access rights to the infrastructure account are restricted.
