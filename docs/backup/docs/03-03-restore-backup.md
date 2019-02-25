---
title: Restore a Kyma cluster
type: Details
---

Restoring a kyma cluster requieres a new / empty installation of kyma. As soon as the cluster is up and running ark can be instructed to start the restore. It is important to restore the system namespaces as well as the user namespaces at the same time, to make sure both are in sync.

Use the following kubectl command to list available backups:

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

To trigger the restore process, create the resource in Kubernetes:

```$ kubectl apply -f <filename>```

To check the restore progress, run the following command:

```$ kubectl describe restore -n heptio-ark  <restore name>```

To validate the result of the restore use `kubectl`. Even the restore is marked completed it may take some time, till all resources are available again.