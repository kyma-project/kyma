# Ark

## Overview

Ark is a tool to back up and restore Kubernetes resources and persistent volumes. It can create backups on demand or on schedule, filter objects which should be backed up, and set TTL (time to live) for stored backups. For more details, see the official [Ark documentation](https://heptio.github.io/velero/v0.9.0/).

## Details

The Ark installation contains only the configuration of storage providers. The configuration of backup content and scope is part of the Backup Resource. To respect the specific architecture of Ark, kyma is delivering  tested sample files as part of the documentation which can be used during the backup process. (See related [kyma documentation](https://kyma-project.io/docs/components/backup)). It is important, that all components which are storing data are part of the sample file configuration to enable administrators to setup a proper backup process.

## Add Components to the Backup

All kubernetes resources in the user namespaces are backuped by default. If there is something not covered by the backup or data which is stored in the system namespaces it has to be added to the sample file in the backup documentation (`/docs/backup/docs/assets`). This sample files are used as a configuration for testing. Details of the configuration attributes are part of the [ark documentation](https://github.com/heptio/velero/blob/master/docs/api-types/backup.md).

If arks functionality is not sufficient it can be extended using [plugins](https://heptio.github.io/velero/v0.10.0/plugins) and [hooks](https://heptio.github.io/velero/v0.10.0/hooks). Plugins are extending ark withouth being part of the binary. Ark Plugins in kyma are stored in the tools section (`tools/ark-plugins`). Hooks are commands executed inside containers and pods during backup. They are configured as part of the backup configuration.

## E2E Testing

The E2E Test for Backup (`tests/end-to-end/backup-restore-test`) is running daily on prow and validating if all components can be restored like expected.

To add components to the backup pipeline it is required to implement a simple go interface.

```go
type BackupTest interface {
    CreateResources(namespace string)
    TestResources(namespace string)
}
```

The `CreateResources` Function is called before the backup to install all required test data. The `TestResources` Function is called after CreateResources to validate if the test data is working like expected. After the pipeline did a backup and restore on the cluster the `TestResources` function is called again to validate the restore was working as expected.

The test has to be registered in the central E2E test.