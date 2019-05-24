---
title: Configuration
type: Details
---

The Velero configuration consists of two custom resources:

- [VolumeSnapshotLocation](https://velero.io/docs/v1.0.0/api-types/volumesnapshotlocation/) CR defines the provider(e.g. aws, gcp or azure) of physical volume snapshots.
- [BackupStorageLocation](https://velero.io/docs/v1.0.0/api-types/backupstoragelocation/) CR defines a bucket or storage location for cluster resources.

A sample BackupStorageLocation CR looks like this:

```apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: default
  namespace: kyma-backup
spec:
  config:
    resourceGroup: BackupStorage
    storageAccount: foo...
  objectStorage:
    bucket: bucket
  provider: azure
```

A sample VolumeSnapshotLocation CR looks like this:

```apiVersion: velero.io/v1
kind: VolumeSnapshotLocation
metadata:
  name: azure-default
  namespace: kyma-backup
spec:
  config:
    apiTimeout: 15m
  provider: azure
```

A Kyma installation provides a set of default snapshot and storage locations. If needed, you can add custom locations in the `kyma-backup` Namespace.