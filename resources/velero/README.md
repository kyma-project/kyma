# Velero

## Overview

Velero is a tool to back up and restore Kubernetes resources and persistent volumes. It can create backups on demand or on schedule, filter objects which should be backed up, and set TTL (time to live) for stored backups. For more details, see the official [Velero documentation](https://velero.io/docs/v1.0.0/).

## Required parameters

This table lists the required parameters of this chart, their descriptions, and default values:

Parameter | Description | Default | Required
--- | --- | --- | ---
`configuration.provider` | The name of the cloud provider where you are deploying velero to (`aws`, `azure`, `gcp`) | none | yes
`configuration.backupStorageLocation.name` | The name of the cloud provider that will be used to actually store the backups (`aws`, `azure`, `gcp`) | none | yes
`configuration.backupStorageLocation.bucket` | The storage bucket where backups are to be uploaded | none | yes
`configuration.backupStorageLocation.config.region` | The cloud provider region (AWS only) | none | yes, if using AWS
`configuration.backupStorageLocation.config.resourceGroup` | The resource group containing the storage account (Azure only) | none | yes, if using Azure
`configuration.backupStorageLocation.config.storageAccount` | The storage account containing the blob container (Azure only) | none | yes, if using Azure
`configuration.volumeSnapshotLocation.name` | The name of the cloud provider the cluster is using for persistent volumes, if any | none | yes, if using PV snapshots
`configuration.volumeSnapshotLocation.config.region` | The cloud provider region (AWS only) | none | yes, if using AWS
`configuration.volumeSnapshotLocation.config.apitimeout` | The API timeout (Azure only) | none | yes, if using Azure
`credentials.useSecret` | Whether a secret should be used for IAM credentials. Set this to `false` when using `kube2iam` | `true` | yes
`credentials.existingSecret` | If specified and `useSecret` is `true`, uses an existing secret with this name instead of creating one | none | yes, if `useSecret` is `true` and `secretContents` is empty
`credentials.secretContents` | If specified and `useSecret` is `true`, contents for the credentials secret | none | yes, if `useSecret` is `true` and `existingSecret` is empty

## Details

The Velero installation contains the configuration for storage. To define the backup content and the scope, Backup custom resources are used. Kyma delivers a tested sample file you can use to run the [backup process](https://github.com/kyma-project/kyma/blob/master/docs/backup/01-01-backup.md). This sample file includes all the Kubernetes resources by default. Modify this file according to your backup needs to allow administrators to set up a proper backup process.

### Velero CLI
Once the Velero server is up and running, you can use the client to interact with it.
1. Download the CLI tool from [here](https://github.com/heptio/velero/releases) based on the `appVersion` in [Chart.yaml](Chart.yaml)
2. Untar using:
```
tar -xvf velero-v<version>-darwin-amd64.tar.gz -C velero-client
```

### Extending Velero

Velero's functionalities can also be extended by using [plugins](https://velero.io/docs/v1.0.0/plugins/) and [hooks](https://velero.io/docs/v1.0.0/hooks/). Velero plugins are added to the Velero server Pod as init containers, and extend Velero without being a part of the binary. Hooks are commands executed inside containers and Pods during the backup process.

### End-to-end tests

For details on end-to-end tests for backup and restore, see [this](../../tests/end-to-end/backup-restore-test/README.md) document.
