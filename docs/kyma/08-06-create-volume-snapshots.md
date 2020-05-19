---
title: Create on-demand volume snapshots
type: Tutorials
---

This tutorial shows how to create on-demand [volume snapshots](https://kubernetes.io/docs/concepts/storage/volume-snapshots/) you can use to provision a new volume or restore an existing one. 

### Steps

Perform the steps:

1. Assume that you have a `pvc-to-backup` PersistentVolumeClaim which you have created using a CSI-enabled StorageClass. Trigger a snapshot by creating a VolumeSnapshot object:

>**NOTE:** You must use CSI-enabled StorageClass to create a PVC, otherwise it won't be backed up.

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

2. Recreate the PVC using the snapshot as the data source:

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

This will create a new `pvc-restored` PVC with pre-populated data from the snapshot. 

