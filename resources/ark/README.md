# Ark

## Overview

Ark is a tool to back up and restore Kubernetes resources and persistent volumes. It can create backups on demand or on schedule, filter objects which should be backed up, and set TTL (time to live) for stored backups. For more details, see the official [Ark documentation](https://heptio.github.io/velero/v0.9.0/).

## Details

The Ark installation contains only the configuration for storage providers. The configuration of backup content and scope is defined in the Backup resource. To comply with the specific Ark architecture, Kyma delivers tested sample files you can use to run the backup process. Add all components which store data in this configuration to allow administrators to set up a proper backup process.

### Add components to backup

All Kubernetes resources in the user Namespaces are backed up by default. However, you must add all data this backup does not cover, or data stored in the system Namespaces, to the configuration files. These serve as a testing configuration. For details on configuration attributes, see [ark documentation](https://github.com/heptio/velero/blob/master/docs/api-types/backup.md).

If Ark's functionality is not sufficient, you can extend it using [plugins](https://heptio.github.io/velero/v0.10.0/plugins) and [hooks](https://heptio.github.io/velero/v0.10.0/hooks). Ark plugins are stored in [`tools`](tools/ark-plugins) directory, and extend Ark without being a part of the binary. Hooks are commands executed inside containers and Pods during the backup process. You can define them in your backup configuration.

## E2E tests

The [E2E test for backup](https://github.com/kyma-project/kyma/tree/master/tests/end-to-end/backup-restore-test) runs daily on Prow and validates if the restore process works for all components as expected.

To add components to the backup pipeline, implement the following go interface:

```go
type BackupTest interface {
    CreateResources(namespace string)
    TestResources(namespace string)
}
```

- The `CreateResources` function is called before the backup to install all required test data.
- The `TestResources` function is called after the `CreateResources` function to validate if the test data is working like expected. After the pipeline did a backup and restore on the cluster the `TestResources` function is called again to validate the restore was working as expected.

Register the test in the [E2E tests](https://github.com/kyma-project/kyma/tree/master/tests/end-to-end/backup-restore-test).