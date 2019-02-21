---
title: How to Restore a Kyma Cluster
---

The restore of a kyma cluster requieres a new / empty installation of kyma. As soon as the cluster is up and running ark can be instructed to start the restore. It is important to restore the system namespaces as well as the user namespaces at the same time, to make sure both are in sync.

A list of available backups can be listed using the kubectl command:

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
    backupName: null // specify to restore a specific backup
    scheduleName: kyma-backup // Applies only if no backup is specified.
    restorePVs: true
    includeClusterResources: true
---
apiVersion: ark.heptio.com/v1
kind: Restore
metadata:
  name: kyma-system-restore
  namespace: heptio-ark
spec:
    backupName: null // specify to restore a specific backup
    scheduleName: kyma-system-backup // Applies only if no backup is specified.
    restorePVs: true
    includeClusterResources: true
```

The restore will be triggered if the resource is created on the kubernetes system:

```$ kubectl apply -f <filename>```

The Progress of the backup can be displayed using `kubectl`:

```$ kubectl describe restore -n heptio-ark  <restore name>```

To validate the result of the restore use `kubectl`. Even the restore is marked completed it can take some time, till a resource is available again.