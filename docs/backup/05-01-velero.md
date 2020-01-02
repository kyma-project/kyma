---
title: Velero chart
type: Configuration
---

This document lists and describes the parameters that you can configure in Velero, split into required and optional parameters.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Required parameters

This table lists the parameters required for Velero to work, their descriptions, and default values:

| Parameter | Description | Default value | Required |
|-----------|-------------|:---------------:|:---------------:|
**configuration.provider** | Specifies the name of the cloud provider where you are deploying Velero to, such as `aws`, `azure`, `gcp`.| None | Yes
**configuration.backupStorageLocation.name** | Specifies the name of the cloud provider used to store backups, such as `aws`, `gcp`, or `azure`. | None | Yes
**configuration.backupStorageLocation.bucket** | Specifies the storage bucket where backups are uploaded. | None | Yes
**configuration.backupStorageLocation.config.region** | Provides the region in which the bucket is created. It only applies to AWS. | None | Yes, if using AWS
**configuration.backupStorageLocation.config.resourceGroup** | Specifies the name of the resource group which contains the storage account for the backup storage location. It only applies to Azure. | none | yes, if using Azure
**configuration.backupStorageLocation.config.storageAccount** | Provides the name of the storage account for the backup storage location. It only applies to Azure.| None | Yes, if using Azure
**configuration.volumeSnapshotLocation.name** | Specifies the name of the cloud provider the cluster is using for persistent volumes. | None | Yes, if using PV snapshots
**configuration.volumeSnapshotLocation.config.region** | Provides the region in which the bucket is created. It only applies to AWS.| None | Yes, if using AWS
**configuration.volumeSnapshotLocation.config.apitimeout** | Defines the amount of time after which an API request returns a timeout status. It only applies to Azure. | None | Yes, if using Azure
**credentials.useSecret** | Specifies if a secret is required for IAM credentials. Set this to `false` when using `kube2iam`. | `true` | Yes
**credentials.existingSecret** | If specified and **useSecret** is `true`, uses an existing secret with this name instead of creating one. | None | Yes, if **useSecret** is `true` and **secretContents** is empty
**credentials.secretContents** | If specified and **useSecret** is `true`, provides the content for the credentials secret. | None | Yes, if **useSecret** is `true` and **existingSecret** is empty
**initContainers.pluginContainer.image** | Provides the image for the respective cloud provider plugin. | `velero/velero-plugin-for-gcp:v1.0.0` | yes, set `velero/velero-plugin-for-microsoft-azure:v1.0.0` for Azure and `velero/velero-plugin-for-aws:v1.0.0` for AWS. See [supported providers](https://velero.io/docs/v1.2.0/supported-providers) for more details

## Configurable parameters

This table lists the non-required configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **schedules** | Sets up a scheduled backup. By default, a scheduled backup runs at 07:00 daily on Monday through Friday. | `scheduled-backup` |
| **configuration.volumeSnapshotLocation.bucket** | Specifies the name of the storage bucket where volume snapshots are uploaded. | None |
| **configuration.backupStorageLocation.prefix** | Specifies the directory inside a storage bucket where backups are located. | None |
| **configuration.backupStorageLocation.config.resourceGroup** | Specifies the name of the resource group which contains the storage account for the backup storage location. It only applies to Azure. | None |
| **configuration.backupStorageLocation.config.s3ForcePathStyle** | Specifies whether to force path style URLs for S3 objects. Set it to `true` if you use a local storage service like MinIO. It only applies to AWS. | None |
| **configuration.backupStorageLocation.config.s3Url** | Specifies the AWS S3 URL. If not provided, Velero generates it from **region** and **bucket**. Use this field for local storage services like MinIO. | None |
| **configuration.backupStorageLocation.config.kmsKeyId** | Specifies the AWS KMS key ID or alias to enable encryption of the backups stored in S3. It only works with AWS S3 and may require explicitly granting key usage rights. | None |
| **configuration.backupStorageLocation.config.publicUrl** | Specifies the parameter used instead of **3Url** when generating download URLs, for example for logs. Use this field for local storage services like MinIO. | None |

See the official Velero documentation for examples and the full list of [parameters](../../resources/velero/README.md), as well as for [VolumeSnapshotLocation](https://velero.io/docs/v1.2.0/api-types/volumesnapshotlocation/) and [BackupStorageLocation](https://velero.io/docs/v1.2.0/api-types/backupstoragelocation/).
