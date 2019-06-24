---
title: Restore a Kyma cluster
type: Details
---

Restoring a Kyma cluster requires a fresh Kyma installation with the same version you want to restore with. As soon as the cluster is up and running, instruct Velero to start the restore process. Restore the system [backup](assets/system-backup.yaml) followed by user Namespaces [backup](assets/all-backup.yaml).

Use this command to list available backups:

```
kubectl get backups -n kyma-backup
```

Sample restore configuration:

```yaml
---
apiVersion: velero.io/v1
kind: Restore
metadata:
  name: kyma-restore
  namespace: kyma-backup
spec:
    backupName: kyma-backup # specify to restore a specific backup
    scheduleName: kyma-backup # Applies only if no backup is specified.
    restorePVs: true
    includeClusterResources: true
---
apiVersion: velero.io/v1
kind: Restore
metadata:
  name: kyma-system-restore
  namespace: kyma-backup
spec:
    backupName: kyma-backup # specify to restore a specific backup
    scheduleName: kyma-system-backup # Applies only if no backup is specified.
    restorePVs: true
    includeClusterResources: true
```

To trigger the restore process, run this command:

```
kubectl apply -f {filename}
```


To check the restore progress, run this command:

```
kubectl describe restore -n kyma-backup {restore name}
```

To validate the result of the restore use the `kubectl get` command.

> **NOTE:** Even if the restore process is complete, it may take some time for the resources to become available again.

> **NOTE:** In order to make Prometheus work after restore following steps need to be done:


```
### Save the prometheus resource in a file
kubectl get Prometheus -n kyma-system monitoring -oyaml --export > prom.yaml

### Delete metadata.generation and metadata.annotation["kubectl.kubernetes.io/last-applied-configuration"]
sed -i prom.yaml '/last-applied-configuration/d;/generation/d;/selfLink/d' prom.yaml

### Reapply the prometheus resource using the file
kubectl -n kyma-system apply -f prom.yaml

```
