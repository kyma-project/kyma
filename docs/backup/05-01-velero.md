---
title: Velero chart
type: Configuration
---

This document lists and describes the parameters that you can configure in velero, split into required and optional parameters.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Required parameters

This table lists the required parameters for velero to work, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **configuration.provider** | The name of the cloud provider where you are deploying velero to (`aws`, `azure`, `gcp`) | None |
| **configuration.backupStorageLocation.name** | The name of the cloud provider that will be used to actually store the backups (`aws`, `azure`, `gcp`) | None |
| **configuration.backupStorageLocation.bucket** | The storage bucket where backups are to be uploaded | None |
| **configuration.backupStorageLocation.config.region** | The cloud provider region (AWS only) | None |
| **configuration.backupStorageLocation.config.resourceGroup** | The resource group containing the storage account (Azure only) | None |
| **configuration.backupStorageLocation.config.storageAccount** | The storage account containing the blob container (Azure only) | None |
| **configuration.volumeSnapshotLocation.name** | The name of the cloud provider the cluster is using for persistent volumes, if any (Only if using PV snapshots) | None |
| **configuration.volumeSnapshotLocation.config.region** | The cloud provider region (AWS only) | None |
| **configuration.volumeSnapshotLocation.config.apitimeout** | The API timeout (Azure only) | None |
| **credentials.useSecret** | Whether a secret should be used for IAM credentials. Set this to `false` when using `kube2iam` | `true` |
| **credentials.existingSecret** | If specified and `useSecret` is `true`, uses an existing secret with this name instead of creating one (Only if `secretContents` is empty) | None |
| **credentials.secretContents** | If specified and `useSecret` is `true`, contents for the credentials secret (Only if `existingSecret` is empty) | None |


## Configurable parameters

This table lists the non-required configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.volumeSnapshotLocation.bucket** | Specifies the name of the storage bucket where volume snapshots are uploaded. | None |
| **global.backupStorageLocation.prefix** | Specifies the directory inside a storage bucket where backups are located. | None |
| **global.backupStorageLocation.config.resourceGroup** | Specifies the name of the resource group which contains the storage account for the backup storage location. It only applies to Azure. | None |
| **global.backupStorageLocation.config.s3ForcePathStyle** | Specifies whether to force path style URLs for S3 objects. Set it to `true` if you use a local storage service like Minio. It only applies to AWS. | None |
| **global.backupStorageLocation.config.s3Url** | Specifies the AWS S3 URL. If not provided, Velero generates it from **region** and **bucket**. Use this field for local storage services like Minio. | None |
| **global.backupStorageLocation.config.kmsKeyId** | Specifies the AWS KMS key ID or alias to enable encryption of the backups stored in S3. It only works with AWS S3 and may require explicitly granting key usage rights. | None |
| **global.backupStorageLocation.config.publicUrl** | Specifies the parameter used instead of **3Url** when generating download URLs, for example for logs. Use this field for local storage services like Minio. | None |

See the official Velero documentation for examples and the full list of [parameters](https://github.com/helm/charts/tree/master/stable/velero#configuration) for [VolumeSnapshotLocation](https://velero.io/docs/v1.0.0/api-types/volumesnapshotlocation/) and [BackupStorageLocation](https://velero.io/docs/v1.0.0/api-types/backupstoragelocation/).
