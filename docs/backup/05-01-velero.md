---
title: Velero chart
type: Configuration
---

To configure the Velero chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.volumeSnapshotLocation** | The location to store volume snapshots created for the backup. It is cloud provider-specific. For example, to set up `aws` as the provider of the snapshot storage location, override the value for the **global.volumeSnapshotLocation.spec.provider** parameter. [See](https://velero.io/docs/v1.0.0/api-types/volumesnapshotlocation/) the official Velero documentation for the full list of configurable parameters and examples. | None |
| **global.backupStorageLocation** |  The location to store backups. It is cloud provider-specific. For example, to set up `aws` as the provider of the backup storage location, override the value for the **global.backupStorageLocation.spec.provider** parameter. [See](https://velero.io/docs/v1.0.0/api-types/backupstoragelocation/) the official Velero documentation for the full list of configurable parameters and examples. | None |
