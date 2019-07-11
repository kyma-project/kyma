---
title: Back up a Kyma cluster
type: Details
---
Kyma provides two validated sample backup specification files:

- [System Namespace Backup](./assets/system-backup.yaml)
- [User Namespace Backup](./assets/all-backup.yaml)


Integrate these files with your scheduled or on-demand configurations to back up system or user Namespaces.

>**NOTE:** To fully back up a cluster, you must back up both user and system Namespaces.

Modify the files to adjust the backup scope. For details about the file format, see the [documentation](https://velero.io/docs/v1.0.0/output-file-format/).

## Create manual backups

If you want to use sample backup configurations, you can use Backup custom resources. Add the following two CRs to the `kyma-backup` Namespace to instruct the Velero server to create a backup. Make sure the indentation is correct.

A sample backup configuration looks like this:

```yaml
---
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: kyma-system-backup
  namespace: kyma-backup
spec:
    {INCLUDE CONTENT OF SYSTEM NAMESPACE BACKUP FILE HERE} ### E.g. docs/backup/assets/system-backup.yaml
---
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: kyma-backup
  namespace: kyma-backup
spec:
    {INCLUDE CONTENT OF USER NAMESPACE BACKUP FILE HERE} ### E.g. docs/backup/assets/all-backup.yaml
```

To create the backup, run the following command:

```
kubectl apply -f {filename}
```

## Schedule periodic backups

If you want to use sample backup configurations, you can use Schedule custom resources. Create two Schedule custom resources in the `kyma-backup` Namespace to instruct the Velero Server to schedule a cluster backup. Make sure the indentation is correct.

A sample scheduled backup configuration looks like this:

```yaml
---
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: kyma-system-backup
  namespace: kyma-backup
spec:
    template:
        {INCLUDE CONTENT OF SYSTEM NAMESPACE BACKUP SPEC HERE}
    schedule: 0 1 * * *
---
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: kyma-backup
  namespace: kyma-backup
spec:
    template:
        {INCLUDE CONTENT OF SYSTEM NAMESPACE BACKUP SPEC HERE}
    schedule: 0 1 * * *
```

To schedule a backup, run the following command:

```
kubectl apply -f {filename}
```

## Backup retention period

To set the retention period of a backup, define the **ttl** parameter in the Backup specification [definition](https://velero.io/docs/v1.0.0/output-file-format/):

```  The amount of time before this backup is eligible for garbage collection.
  ttl: 24h0m0s
  ```
