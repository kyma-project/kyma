---
title: Back up a Kyma cluster
type: Tutorial
---
Follow this tutorial to learn how to use the [backup.yaml](./assets/backup.yaml) specification file to create a manual or scheduled Kyma backup. For details about the file format, see [this](https://velero.io/docs/v1.2.0/api-types/backup/) document.

## Prerequisites

Install Velero using [these](/components/backup/#installation-install-velero) instructions. 

## Steps

Follow the steps to back up Kyma.

<div tabs name="backup">
  <details>
  <summary label="manual-backup">
  Manual backup
  </summary>

To create a manual backup, use the Backup custom resource based on Velero's [Backup](https://velero.io/docs/v1.2.0/api-types/backup/) API type. Deploy the following CR to the `kyma-system` Namespace to instruct the Velero server to create a backup. Make sure the indentation is correct.

A sample backup configuration looks like this:

```yaml
---
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: kyma-backup
  namespace: kyma-system
spec:
  includedNamespaces:
  - '*'
  includedResources:
  - '*'
  includeClusterResources: true
  storageLocation: default
  volumeSnapshotLocations:
  - default
```

To trigger the backup process, run the following command:

```
kubectl apply -f {filename}
```
  </details>
  <details>
  <summary label="scheduled-backup">
  Scheduled backup
  </summary>

By default, the backup runs once a day every day from Monday to Friday. To set up a different backup schedule, create a Schedule custom resource based on the Velero's [Backup](https://velero.io/docs/v1.2.0/api-types/backup/) API type. Deploy it in the `kyma-system` Namespace to instruct the Velero Server to schedule a cluster backup. Make sure the indentation is correct.

A sample scheduled backup configuration looks like this:

```yaml
---
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: kyma-backup
  namespace: kyma-system
spec:
  template:
    includedNamespaces:
    - '*'
    includedResources:
    - '*'
    includeClusterResources: true
    storageLocation: default
    volumeSnapshotLocations:
    - default
  schedule: 0 1 * * *
```

To schedule a backup, run the following command:

```bash
kubectl apply -f {filename}
```

  </details>
</div>

## Backup retention period

To set the retention period of a backup, define the **ttl** parameter in the Backup specification [definition](https://velero.io/docs/v1.2.0/api-types/backup/):

```yaml  
# The amount of time before this backup is eligible for garbage collection.
ttl: 24h0m0s
```
