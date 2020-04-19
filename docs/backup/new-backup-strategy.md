# Kyma Backup Strategy

User load on a Kyma cluster typically consists of various Kubernetes objects and volumes. Kyma relies on the backing cloud provider for periodic backups of the Kubernetes objects. That's why it does not require the user to do any manual settings to take backups.

For example, Gardener uses etcd as the Kubernetes' backing store for all cluster data. That is, all Kubernetes objects are stored on etcd. Gardener has perodic jobs to take major and minor backups of etcd database. A major backup including all the resources takes place every hour, and each minor backups including only the changes inbetween take place every five minutes. In case of a problem on etcd database, Gardener automatically restores the Kubernetes cluster using the latest backup.

However, volumes are typically not a part of these backups. That's why Kyma encourages users to take periodic backups of their volumes. This can be done using the new Kubernetes API `VolumeSnapshot` that is explained below.

## On-Demand Volume Snapshots

Kubernetes provides a new API resource called VolumeSnapshot that can be used to take the snapshot of a volume on Kubernetes. A snapshot can be used either to provision a new volume (pre-populated with the snapshot data) or to restore the existing volume to a previous state (represented by the snapshot).

VolumeSnapshot support is only available for CSI drivers. However, not all the CSI drivers support the volume snapshot functionality. You can find a list of all the drivers with the supported functionalities [here](https://kubernetes-csi.github.io/docs/drivers.html).

An example VolumeSnapshot created to take a snapshot from the PVC named `pvc-to-backup`:

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: volume-snapshot
spec:
  volumeSnapshotClassName: csi-snapshot-class
  source:
    persistentVolumeClaimName: pvc-to-backup
```

An example PVC using this snapshot as the data source:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-restored
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: csi-storage-class
  resources:
    requests:
      storage: 10Gi
  dataSource:
    name: volume-snapshot
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
```

You can find more information about VolumeSnapshots [here](https://kubernetes.io/docs/concepts/storage/volume-snapshots/).

Currently none of the cloud providers support this API out-of-the-box yet. You must follow the instructions below to enable this feature for various providers.

### AKS

Minimum Kubernetes version supported is 1.17.

Install CSI driver following [this documentation](https://github.com/kubernetes-sigs/azuredisk-csi-driver/blob/master/docs/install-csi-driver-master.md).

Then, you can follow [this example](https://github.com/kubernetes-sigs/azuredisk-csi-driver/tree/master/deploy/example/snapshot) to see how you can create a VolumeSnapshot.

### GKE

Minimum Kubernetes version supported is 1.14.

Enable the required feature gate on the cluster following [this document](https://cloud.google.com/kubernetes-engine/docs/how-to/gce-pd-csi-driver#enabling_on_a_new_cluster).

Install CSI driver following the instructions [here](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/blob/master/docs/kubernetes/user-guides/snapshots.md#kubernetes-snapshots-user-guide-alpha).

Then, you can follow [this example](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver/blob/master/docs/kubernetes/user-guides/snapshots.md#snapshot-example) to see how you can create a VolumeSnapshot.

### Gardener Azure

Gardener Azure does not currently support CSI drivers, that's why VolumeSnapshots cannot be used. Its support is planned for Kubernetes 1.19 [#3](https://github.com/gardener/gardener-extension-provider-azure/issues/3).

### Troubleshooting

#### VolumeSnapshot creation failed

If a PersistentVolumeClaim is not bound, the attempt to create a volume snapshot from that PersistentVolumeClaim will fail. No retries will be attempted. An event will be logged to indicate that the PersistentVolumeClaim is not bound.

Note that this could happen if the PersistentVolumeClaim spec and the VolumeSnapshot spec are in the same YAML file. In this case, when the VolumeSnapshot object is created, the PersistentVolumeClaim object is created but volume creation is not complete and therefore the PersistentVolumeClaim is not yet bound. You must wait until the PersistentVolumeClaim is bound and then create the snapshot.
