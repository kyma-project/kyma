# Backup setup

## Overview

Backup setup is used to configure the `spec:` values for ark server to execute the backup in Kyma.

* Include namespaces
* Exclude namespaces
* Include resources
* Exclude resources

## Usage

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

### Apply backup.yaml

`kondemandatx apply -f tools/backup-setup/config/backup.yaml`