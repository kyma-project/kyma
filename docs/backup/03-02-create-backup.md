---
title: Back up a Kyma cluster
type: Details
---
Kyma provides a validated sample [Backup.yaml](./assets/backup.yaml) specification file. Integrate it with your scheduled or on-demand configurations to back up Kyma. Modify the file to adjust the backup scope. For details about the file format, see this [documentation](https://velero.io/docs/v1.0.0/api-types/backup/).

## Create manual backups

To create manual backups, use Backup custom resources. Deploy the following CR to the `kyma-backup` Namespace to instruct the Velero server to create a backup. Make sure the indentation is correct.

A sample backup configuration looks like this:

```yaml
---
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: kyma-backup
  namespace: kyma-backup
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

## Schedule periodic backups

To set up periodic backups, use Schedule custom resources. Deploy Schedule custom resources in the `kyma-backup` Namespace to instruct the Velero Server to schedule a cluster backup. Make sure the indentation is correct.

A sample scheduled backup configuration looks like this:

```yaml
---
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: kyma-backup
  namespace: kyma-backup
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

## Backup retention period

To set the retention period of a backup, define the **ttl** parameter in the Backup specification [definition](https://velero.io/docs/v1.0.0/api-types/backup/):

```yaml  
# The amount of time before this backup is eligible for garbage collection.
ttl: 24h0m0s
```
