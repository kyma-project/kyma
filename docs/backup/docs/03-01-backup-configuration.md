---
title: Configuration
type: Details
---

The Ark configuration consists of two custom resources:

- [VolumeSnapshotLocation](https://github.com/heptio/velero/blob/master/docs/api-types/volumesnapshotlocation.md) CR defines the provider of physical volume snapshots.
- [BackupStorageLocation](https://github.com/heptio/velero/blob/master/docs/api-types/backupstoragelocation.md) CR defines a bucket or storage location for cluster resources.

A sample BackupStorageLocation CR looks like this:

```apiVersion: ark.heptio.com/v1
kind: BackupStorageLocation
metadata:
  name: default
  namespace: heptio-ark
spec:
  config:
    resourceGroup: BackupStorage
    storageAccount: foo...
  objectStorage:
    bucket: bucket
  provider: azure
```

A sample VolumeSnapshotLocation CR looks like this:

```apiVersion: ark.heptio.com/v1
kind: VolumeSnapshotLocation
metadata:
  name: azure-default
  namespace: heptio-ark
spec:
  config:
    apiTimeout: 15m
  provider: azure
```

A Kyma installation provides a set of default snapshot and storage locations. If needed, you can add custom locations in the `heptio-ark` Namespace.