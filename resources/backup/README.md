# Backup

## Overview

Velero is the tool used to back up and restore Kubernetes resources and persistent volumes. It can create on-demand or scheduled backups, filter objects which should be backed up, and set TTL (time to live) for stored backups. For more details, see the official [Velero documentation](https://velero.io/docs/v1.2.0/).

## Required parameters

This table lists the required parameters of this chart, their descriptions, and default values:

Parameter | Description | Default | Required
--- | --- | --- | ---
**configuration.provider** | Specifies the name of the cloud provider where you are deploying Velero to, such as `aws`, `azure`, `gcp`.| None | yes
**configuration.backupStorageLocation.name** | Specifies the name of the cloud provider used to store backups, such as `aws`, `gcp`, or `azure`. | None | yes
**configuration.backupStorageLocation.bucket** | Specifies the storage bucket where backups are uploaded. | None | yes
**configuration.backupStorageLocation.config.region** | Provides the region in which the bucket is created. It only applies to AWS. | None | yes, if using AWS
**configuration.backupStorageLocation.config.resourceGroup** | Specifies the name of the resource group which contains the storage account for the backup storage location. It only applies to Azure. | None | yes, if using Azure
**configuration.backupStorageLocation.config.storageAccount** | Provides the name of the storage account for the backup storage location. It only applies to Azure.| None | yes, if using Azure
**configuration.volumeSnapshotLocation.name** | Specifies the name of the cloud provider the cluster is using for persistent volumes. | None | yes, if using PV snapshots
**configuration.volumeSnapshotLocation.config.region** | Provides the region in which the bucket is created. It only applies to AWS.| None | yes, if using AWS
**configuration.volumeSnapshotLocation.config.apitimeout** | Defines the amount of time after which an API request returns a timeout status. It only applies to Azure. | None | yes, if using Azure
**credentials.useSecret** | Specifies if a secret is required for IAM credentials. Set this to `false` when using `kube2iam`. | `true` | yes
**credentials.existingSecret** | If specified and `useSecret` is `true`, uses an existing secret with this name instead of creating one. | None | yes, if `useSecret` is `true` and `secretContents` is empty
**credentials.secretContents** | If specified and `useSecret` is `true`, provides the content for the credentials secret. | None | yes, if `useSecret` is `true` and `existingSecret` is empty
**initContainers[0].image** | Provides the image for the respective cloud provider plugin. | `velero/velero-plugin-for-gcp:v1.0.0` | yes, set `velero/velero-plugin-for-microsoft-azure:v1.0.0` for Azure and `velero/velero-plugin-for-aws:v1.0.0` for AWS. See https://velero.io/docs/v1.2.0/supported-providers/ for more details

## Details

The Velero installation contains storage configuration. You can use the Backup custom resources to define the backup content and the scope. Kyma comes with a tested sample file you can use to run the [backup process](https://github.com/kyma-project/kyma/blob/master/docs/backup/01-01-backup.md). This sample file includes all the Kubernetes resources by default. Modify this file according to your backup needs to allow the administrators to set up a proper backup process.

### Velero CLI
Once the Velero server is up and running, you can use the client to interact with it.
1. Download the CLI tool from [here](https://github.com/heptio/velero/releases) based on the `appVersion` in [Chart.yaml](Chart.yaml)
2. Untar using:
```
tar -xvf velero-v<version>-darwin-amd64.tar.gz -C velero-client
```

### Extending Velero

You can extend Velero functionality using [plugins](https://velero.io/docs/v1.2.0/overview-plugins/) and [hooks](https://velero.io/docs/v1.2.0/hooks/). Velero plugins are added to the Velero server Pod as init containers and extend Velero without being a part of the binary. Hooks are commands executed inside containers and Pods during the backup process.

Kyma comes with a couple of [plugins](../../components/backup-plugins/) necessary to properly restore the Kyma cluster and its resources.

### End-to-end tests

For details on end-to-end tests for backup and restore, see [this](../../tests/end-to-end/backup-restore-test/README.md) document.
