---
title: Configuration
---

Ark configuration consists of two custom resources:

- The VolumeSnapshotLocation CR defines the provider of physical volume snapshots.
- The BackupStorageLocations CR defines a bucket or storage location for cluster resources.

A Kyma installation provides a set of default snapshot and storage locations. If necessary, you can add custom locations in the `heptio-ark` Namespace.

A sample `BackupStorageLocation` configuration file looks as follows:

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

For a detailed description of all options, see the [Heptio Ark Documentation](https://github.com/heptio/velero/blob/master/docs/api-types/volumesnapshotlocation.md).

A sample `VolumeSnapshotLocation` configuration file looks as follows:

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

A detailed description of all options can be found in the [Heptio Ark Documentation](https://github.com/heptio/velero/blob/master/docs/api-types/volumesnapshotlocation.md).
