---
title: How to Backup a Kyma Cluster
---
Ark provides two ways of backup. Periodically scheduled backups and manual backups. Both backup types including a `BackupSpec` defining which resources to backup.

Kyma is providing two validated sample `BackupSpec` files to backup system and user namespaces. The specs can be integrated into Scheduled backups as well as ad hoc backups. To have a full backup of a cluster, system and user namespaces needs to be backuped.

<!-- TODO: Un comment asson as the resources are available. - [System Namespace Backup](assets/system-backup.yaml)
- [User Namespace Backup](all-backup.yaml) -->

Changing thees files will adjust the scope of the backup. Details about the file format can be found as part of the [Ark Documentation](https://github.com/heptio/velero/blob/master/docs/api-types/backup.md)

## Create Manual Backups

In order to make use of the sample backup configurations it is not possible to use the ark commandline tool to create a backup. To instruct the Ark Server to create a backup two `Backup` Resources has to be created in the `heptio-ark` namespace. Please ensure the indentation is correct.

Sample Backup Configuration:

```yaml
---
apiVersion: ark.heptio.com/v1
kind: Backup
metadata:
  name: kyma-system-backup
  namespace: heptio-ark
spec:
    <INCLUDE CONTENT OF SYSTEM NAMESPACE BACKUP FILE HERE>
---
apiVersion: ark.heptio.com/v1
kind: Backup
metadata:
  name: kyma-backup
  namespace: heptio-ark
spec:
    <INCLUDE CONTENT OF USER NAMESPACE BACKUP FILE HERE>
```

The Backup will be created if the file is uploaded to kubernetes:

```$ kubectl apply -f <filename>```

## Schedule Time Triggeres Backups

In order to make use of the sample backup configurations it is not possible to use the ark commandline tool to schedule a backup. To instruct the Ark Server to schedule a cluster backup two `Schedule` Resources has to be created in the `heptio-ark` namespace. Please ensure the indentation is correct.

```yaml
---
apiVersion: ark.heptio.com/v1
kind: Schedule
metadata:
  name: kyma-system-backup
  namespace: heptio-ark
spec:
    template:
        <INCLUDE CONTENT OF SYSTEM NAMESPACE BACKUP SPEC HERE>
    schedule: 0 1 * * *
---
apiVersion: ark.heptio.com/v1
kind: Schedule
metadata:
  name: kyma-backup
  namespace: heptio-ark
spec:
    template:
        <INCLUDE CONTENT OF SYSTEM NAMESPACE BACKUP SPEC HERE>
    schedule: 0 1 * * *
```

The Backup will be scheduled if the file is uploaded to kubernetes:

```$ kubectl apply -f <filename>```

## Backup Retention Period

The retention period of a backup can be set using the `ttl` attribute in the `BackupSpec` [definition](https://github.com/heptio/velero/blob/master/docs/api-types/backup.md).

```  # The amount of time before this backup is eligible for garbage collection.
  ttl: 24h0m0s```
