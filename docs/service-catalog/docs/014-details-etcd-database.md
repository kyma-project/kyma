---
title: Etcd Database
type: Details
---

## Overview

The Service Catalog requires an `etcd` database cluster for a production use. 
It has a separate `etcd` cluster defined in the Service Catalog [etcd][sc-etcd-sub-chart] sub-chart. 
The [etcd-operator][etcd-operator] installs and manages the `etcd` clusters deployed to Kubernetes,
and also automates tasks related to operating an `etcd` cluster, for example executing [backup][etcd-backups] and [restore][etcd-restores] procedures. 

> **NOTE:** The [etcd-operator][etcd-operator] is Namespace-scoped and is installed only in `kyma-system` Namespace.

## Details

This section describes the backup and restore processes of the `etcd` cluster for the Service Catalog.

### Backup

To execute the backup process, you must set the following values in the [core][core-chart-values] chart:

| Property name              | Description |
|---------------------------------------------------|---|
| **etcd-operator.backupOperator.enabled**            | If set to true, the [etcd-operator][etcd-operator-chart] chart installs the [etcd-backup-operator][etcd-backup-operator-chart]. The etcd-operator also creates the [Secret][abs-creds] with the **storage-account** and **storage-key** keys.  |
| **etcd-operator.backupOperator.abs.storageAccount** | The name of the storage account for the Azure Blob Storage (ABS). It stores the value for the **storage-account** Secret key. |
| **etcd-operator.backupOperator.abs.storageKey**     | The key value of the storage account for the ABS. It stores the value for the **storage-key** Secret key. |
| **global.etcdBackupABS.containerName**              | The ABS container to store the backup. If set, the Service Catalog [sub-chart][sc-backup-sub-chart] installs the CronJob which executes periodically the [Etcd Backup][etcd-backup-app] application. For more information on how to configure the backup CronJob, see the [Etcd Backup][etcd-backup-app-readme] documentation. |

> **NOTE:** If you set the **storageAccount**, **storageKey**, and **containerName** properties, the **etcd-operator.backupOperator.enabled** must be set to `true`. 

### Restore

Follow this instruction to restore an `etcd` cluster from the existing backup.

> **NOTE:** You must have the backup files created by the CronJob backup application from the previous section.

1. Install the etcd-restore-operator:
```bash
kubectl create -f assets/etcd-restore/operator-deploy.yaml
```

2. Create the EtcdRestore Custom Resource Definition:
```bash
kubectl create -f assets/etcd-restore/restore-crd.yaml 
```

3. Export the **ABS_PATH** environment variable with the path to the last successful backup file.
```bash
export ABS_PATH=$(kubectl get cm -n kyma-system sc-recorded-etcd-backup-data -o=jsonpath='{.data.abs-backup-file-path-from-last-success}')
```

> **NOTE:** The ConfigMap name is defined [here][sc-backup-sub-chart] as the **APP_BACKUP_CONFIG_MAP_NAME_FOR_TRACING**.

4. Export the **SECRET_NAME** environment variable with the Secret name to the ABS:
```bash
export SECRET_NAME=etcd-backup-abs-credentials
```

> **NOTE:** The Secret name is defined [here](../../../resources/core/values.yaml) under the **global.etcdBackupABS.secretName** property.

5. Create the EtcdRestore Custom Resource which triggers a restore process:
```bash
sed -e "s|<full-abs-path>|$ABS_PATH|g" \
    -e "s|<abs-secret>|$SECRET_NAME|g" \
    assets/etcd-restore/restore-cr.tpl.yaml \
    | kubectl create -f -
```

Now the etcd-restore-operator restores a new cluster from the backup.

6. See the status of the Pods and wait until all of them are ready. Check the [number][sc-etcd-cluster-no] of the Pods which should be in the`RUNNING` state before going to the next step:
```bash
watch -n 1 kubectl get pod -n kyma-system -l app=etcd,etcd_cluster=service-catalog-core-etcd
```

7. Restart the Service Catalog `apiserver` Pod:
```bash
kubectl delete pod -n kyma-system -l app=core-catalog-apiserver
```

8. Restart the Service Catalog `controller-manager` Pod:
```bash
kubectl delete pod -n kyma-system -l app=core-catalog-controller-manager
```

9. Clean-up the etcd-restore-operator and EtcdRestore CR:
```bash
kubectl delete -f assets/etcd-restore/restore-cr.tpl.yaml
kubectl delete -f assets/etcd-restore/restore-crd.yaml
kubectl delete -f assets/etcd-restore/operator-deploy.yaml
```

[etcd-operator]:https://github.com/coreos/etcd-operator
[etcd-backups]:https://github.com/coreos/etcd-operator/blob/master/doc/user/walkthrough/backup-operator.md
[etcd-restores]:https://github.com/coreos/etcd-operator/blob/master/doc/user/walkthrough/restore-operator.md

[sc-etcd-sub-chart]:../../../resources/core/charts/service-catalog/charts/etcd/templates/etcd-cluster-cr.yaml
[sc-etcd-cluster-no]:../../../resources/core/charts/service-catalog/charts/etcd/templates/etcd-cluster-cr.yaml#L6
[sc-backup-sub-chart]:../../../resources/core/charts/service-catalog/charts/etcd/templates/backup-job.yaml
[etcd-operator-chart]:../../../resources/core/charts/service-catalog/charts/etcd
[etcd-backup-operator-chart]:../../../resources/core/charts/etcd-operator/templates/backup-deployment.yaml
[core-chart-values]:../../../resources/core/values.yaml

[etcd-backup-app-readme]:../../../tools/etcd-backup/README.md
[etcd-backup-app]:../../../tools/etcd-backup

[abs-creds]:../../../resources/core/charts/etcd-operator/templates/etcd-backup-abs-storage-secret.yaml