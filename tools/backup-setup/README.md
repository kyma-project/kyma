# Backup setup

## Overview

Backup setup is used to configure the specifications,`spec:`  for ark server to execute the backup operation in Kyma.

Some of the configurations to apply are:

* Include namespaces
* Exclude namespaces
* Include resources
* Exclude resources


## Usage

Developers and administrators can change [backup.yaml](config/backup.yaml) following the instructions below.

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

### Executing a backup using backup.yaml

In order to execute a backup execute:

`kubectl apply -f tools/backup-setup/config/backup.yaml`

Also it is possible to add labels and change the name of the backup.