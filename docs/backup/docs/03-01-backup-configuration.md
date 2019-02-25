---
title: Configuration
---

The configuration of ark consist of two custom resources: 

- The VolumeSnapshotLocation CR defines the provider of physical volume snapshots.
- The BackupStorageLocations CR defines a bucket or storage location for cluster resources.

A kyma installation is delivered with a set of default snapshot and storage locations. If needed it is possible to add custom locations in the `heptio-ark` namespace.

A sample YAML `BackupStorageLocation` configuration file looks as follows:

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

For a detailed description of all options, see the [Heptio Ark documentation](https://github.com/heptio/velero/blob/master/docs/api-types/backupstoragelocation.md).

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
