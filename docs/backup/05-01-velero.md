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
| **global.volumeSnapshotLocation** | The location to store volume snapshots created for the backup. It is cloud provider-specific. [See](https://velero.io/docs/v1.0.0/api-types/volumesnapshotlocation/) the official Velero documentation for the full list of configurable parameters. | None |
| **global.backupStorageLocation** |  The location to store backups. It is cloud provider-specific. [See](https://velero.io/docs/v1.0.0/api-types/backupstoragelocation/) the official Velero documentation for the full list of configurable parameters. | None |
