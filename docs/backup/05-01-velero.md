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
| **global.volumeSnapshotLocation.name** | Specifies the name of the cloud provider used to store the volume snapshots, such as `aws`, `gcp`, or `azure`. | None |
| **global.volumeSnapshotLocation.bucket** | Specifies the name of the storage bucket where volume snapshots are uploaded. | None |
| **global.volumeSnapshotLocation.config.region** | Provides the region in which the bucket is created. It only applies to AWS. See the full list of [AWS regions](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions). | None |
| **global.volumeSnapshotLocation.config.apiTimeout** | Defines the amount of time after which an API request returns with a timeout status. It only applies to Azure. | None |
| **global.backupStorageLocation.name** | Specifies the name of the cloud provider used to store the backups, such as `aws`, `gcp`, or `azure`. | None |
| **global.backupStorageLocation.bucket** | Specifies the storage bucket where backups are uploaded.| None |
| **global.backupStorageLocation.prefix** | Specifies the directory inside a storage bucket where backups are located. | None |
| **global.backupStorageLocation.config.resourceGroup** | Specifies the name of the resource group containing the storage account for the backup storage location. It only applies to Azure. | None |
| **global.backupStorageLocation.config.storageAccount** | Provides the name of the storage account for the backup storage location. It only applies to Azure. | None |
| **global.backupStorageLocation.config.region** | Provides the region in which the bucket is created. It only applies to AWS. See the full list of [AWS regions](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions). | None |
| **global.backupStorageLocation.config.s3ForcePathStyle** | Specifies whether to force path style URLs for S3 objects.	Set it to `true` if you use a local storage service like Minio. It only applies to AWS. | None |
| **global.backupStorageLocation.config.s3Url** | Specifies the AWS S3 URL. If not provided, Velero generates it from **region** and **bucket**. Use this field for local storage services like Minio. | None |
| **global.backupStorageLocation.config.kmsKeyId** | Specifies an AWS KMS key ID or alias to enable encryption of the backups stored in S3. It only works with AWS S3 and may require explicitly granting key usage rights. | None |
| **global.backupStorageLocation.config.publicUrl** | Specifies the parameter used instead of **3Url** when generating download URLs, for example for logs. Use this field for local storage services like Minio. | None |

See the official Velero documentation for examples and the full list of configurable parameters for [VolumeSnapshotLocation](https://velero.io/docs/v1.0.0/api-types/volumesnapshotlocation/) and [BackupStorageLocation](https://velero.io/docs/v1.0.0/api-types/backupstoragelocation/).
