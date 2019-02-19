# Backup setup

## Overview

Backup setup is used to configure the specifications,`spec:`  for ark server to execute the backup operation in Kyma.

**See:** [Ark in Kyma](resources/ark/README.md) 

Some of the configurations to apply are:

* Include namespaces
* Exclude namespaces
* Include resources
* Exclude resources

Backup setup includes two yaml file for configuring backups:

- [all-backup.yaml](tools/backup-setup/config/all-backup.yaml)

It is meant to be the one for managing critical data backups such as **business** related namespaces as `production`. System namespaces are not included for in this backup.

- [system-backup.yaml](tools/backup-setup/config/system-backup.yaml)

It is intended for system backups such as `kyma-system` and other related namespaces. 

## Usage

Developers and administrators can change either [system-backup.yaml](tools/backup-setup/config/system-backup.yaml) or [system-backup.yaml](tools/backup-setup/config/system-backup.yaml) as it is needed.

The `spec:` section is where the backup is configured. A short explanation can be see it below: 

```yaml
spec:
  # Array of namespaces to include in the backup. If unspecified, all namespaces are included.
  includedNamespaces:
  - '*'
  # Array of namespaces to exclude from the backup.
  excludedNamespaces:
  - kube-system
  # Array of resources to include in the backup. Preferably fully-qualified. 
  includedResources:
  - '*'
  # Array of resources to exclude from the backup. Resources may be shortcuts (e.g. 'po' for 'pods')
  # or fully-qualified. Optional.
  excludedResources:
  - storageclasses.storage.k8s.io
  # Whether or not to include cluster-scoped resources. Valid values are true, false, and
  # null/unset. 
  includeClusterResources: null
  # Individual objects must match this label selector to be included in the backup.
  labelSelector:
    matchLabels:
```

Here is the official description for the [Backup API](https://github.com/heptio/velero/blob/release-0.9/docs/api-types/backup.md) in [Velero project](https://github.com/heptio/velero)

### Executing a backup using backup.yaml

Backups can be easily automated by executing from a shell script or golang binary the kubectl instruction below: 

`kubectl apply -f tools/backup-setup/config/system-backup.yaml`

`kubectl apply -f tools/backup-setup/config/system-backup.yaml`

