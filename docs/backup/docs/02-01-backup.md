---
Title: Configuration
---

The configuration of ark consist of two custom resources. `VolumeSnapshotLocation` are defining the provider of phisical volume snapshots. `BackupStorageLocations` is defining a bucket or location where to store cluster resources. A kyma installation is delivered with a set of default snapshot and storage locations. If needed it is possible to add custom locations in the `heptio-ark` namespace.

A sample YAML `BackupStorageLocation` looks like the following:

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

A detailed description of all options can be found in the [Heptio Ark Documentation](https://github.com/heptio/velero/blob/master/docs/api-types/backupstoragelocation.md).

A sample YAML `BackupStorageLocation` looks like the following:

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
