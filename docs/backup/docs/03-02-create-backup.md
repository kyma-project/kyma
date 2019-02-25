---
title: Back up a Kyma cluster
type: Details
---
Kyma provides two validated sample `BackupSpec` files to back up system and user Namespaces using scheduled or on-demand backups. For a full backup of a cluster, you must back up the system and user Namespaces.

Kyma is providing two validated sample `BackupSpec` files to backup system and user namespaces. The specs can be integrated into Scheduled backups as well as ad hoc backups. To have a full backup of a cluster, system and user namespaces needs to be backuped.

<!-- TODO: Un comment asson as the resources are available. - [System Namespace Backup]assets/system-backup.yaml
- [User Namespace Backup]all-backup.yaml -->

Modifying these files allows you to adjust the scope of the backup. For details about the file format, see the [Ark documentation](https://github.com/heptio/velero/blob/master/docs/api-types/backup.md).

## Create manual backups

If you want to use sample backup configurations, you cannot use the Ark command line tool to create backups. To instruct the Ark Server to create a backup, add the following two Backup resources in the `heptio-ark` Namespace. Make sure the indentation is correct.

Sample backup configuration:

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

To create the backup, upload the backup file to Kubernetes:

```$ kubectl apply -f <filename>```

## Schedule periodic backups

If you want to use sample backup configurations, you cannot use the Ark command line tool to schedule backups. To instruct the Ark Server to schedule a cluster backup, create two Schedule resources in the `heptio-ark` Namespace. Make sure the indentation is correct.

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

To schedule a backup, upload the files to Kubernetes:

```$ kubectl apply -f <filename>```

## Backup retention period

To set the retention period of a backup, define the **ttl** parameter in the `BackupSpec` [definition](https://github.com/heptio/velero/blob/master/docs/api-types/backup.md):

```  # The amount of time before this backup is eligible for garbage collection.
  ttl: 24h0m0s```
