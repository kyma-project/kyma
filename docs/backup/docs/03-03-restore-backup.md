---
title: Restore a Kyma cluster
type: Details
---

Restoring a Kyma cluster requires a fresh Kyma installation. As soon as the cluster is up and running, instruct Ark to start the restore process. Make sure to restore the system and user Namespaces at the same time so both are in sync.

Use this command to list available backups:

```$ kubectl get backups -n heptio-ark```

Sample restore configuration:

```yaml
---
apiVersion: ark.heptio.com/v1
kind: Restore
metadata:
  name: kyma-restore
  namespace: heptio-ark
spec:
    backupName: null # specify to restore a specific backup
    scheduleName: kyma-backup # Applies only if no backup is specified.
    restorePVs: true
    includeClusterResources: true
---
apiVersion: ark.heptio.com/v1
kind: Restore
metadata:
  name: kyma-system-restore
  namespace: heptio-ark
spec:
    backupName: null # specify to restore a specific backup
    scheduleName: kyma-system-backup # Applies only if no backup is specified.
    restorePVs: true
    includeClusterResources: true
```

To trigger the restore process, run this command:

```$ kubectl apply -f <filename>```

To check the restore progress, run this command:

```$ kubectl describe restore -n heptio-ark  <restore name>```

To validate the result of the restore use the `kubectl get` command.

> **NOTE:** Even if the restore process is complete, it may take some time for the resources to become available again.