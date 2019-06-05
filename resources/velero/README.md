# Velero

## Overview

Velero is a tool to back up and restore Kubernetes resources and persistent volumes. It can create backups on demand or on schedule, filter objects which should be backed up, and set TTL (time to live) for stored backups. For more details, see the official [Velero documentation](https://heptio.github.io/velero/v0.11.0/).

## Details

The Velero installation contains only the configuration for storage providers. The Backup custom resource defines the backup content and scope configuration. To comply with the specific Velero architecture, Kyma delivers tested sample files you can use to run the [backup process](https://github.com/kyma-project/kyma/blob/master/docs/backup/01-01-backup.md). Add all components which store data in this configuration to allow administrators to set up a proper backup process.

### Add components to backup

All Kubernetes resources in the user Namespaces are backed up by default. However, you must add all data this backup does not cover, or data stored in the system Namespaces, to the configuration files. These serve as a testing configuration. For details on configuration attributes, see [Velero documentation](https://github.com/heptio/velero/blob/release-0.11/docs/api-types/backup.md).

### Velero CLI
Once the Velero server is up and running, you need the client before you can use it.
1. Download the CLI tool from [here](https://github.com/heptio/velero/releases) based on the `appVersion` in [Chart.yaml](Chart.yaml)
2. Untar using:
```
tar -xvf velero-v<version>-darwin-amd64.tar.gz -C velero-client
```

If Velero's functionality is not sufficient, you can extend it using [plugins](https://heptio.github.io/velero/v0.11.0/plugins) and [hooks](https://heptio.github.io/velero/v0.11.0/hooks). Velero plugins are stored in [`tools`](tools/velero-plugins) directory, and extend Velero without being a part of the binary. Hooks are commands executed inside containers and Pods during the backup process. You can define them in your backup configuration.

## End-to-end tests

For details on end-to-end tests for backup and restore, see [this](../../tests/end-to-end/backup-restore-test/README.md) document.
