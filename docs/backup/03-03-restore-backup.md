---
title: Restore a Kyma cluster
type: Details
---

Restoring a Kyma cluster requires installing Velero. Download Velero CLI and use it to install the Velero server. Then, using the client, instruct Velero to start the restore process. Restore the CRDs, Services, and Endpoints first, and then the rest of the resources.

Download and install [Velero CLI v1.0.0](https://github.com/heptio/velero/releases/tag/v1.0.0).

Install Velero server using the same bucket information where the backups reside:

```
velero install --bucket <BUCKET> --provider <CLOUD_PROVIDER> --secret-file <CREDENTIALS_FILE> --restore-only --wait
```

List available backups:

```
velero get backups
```

Restore Kyma CRDs, Services, and Endpoints:

```
velero restore create --from-backup <BACKUP_NAME> --include-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --include-cluster-resources --wait
```

Restore the rest of Kyma:

```
velero restore create --from-backup <BACKUP_NAME> --exclude-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --include-cluster-resources --restore-volumes --wait
```

Once the status of the restore is `COMPLETED`, verify the health of Kyma by checking the Pods:

> **NOTE:** Even if the restore process is complete, it may take some time for the resources to become available again.

```
kubectl get pods --all-namespaces
```

> **NOTE:** Because of [this issue](https://github.com/heptio/velero/issues/1633) in Velero, Custom Resources are sometimes not properly restored. In this case, you can rerun the second restore command and check if the Custom Resources are restored. For example, run the following command to print several VirtualService Custom Resources:

```
kubectl get virtualservices --all-namespaces
```
