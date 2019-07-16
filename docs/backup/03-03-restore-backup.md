---
title: Restore a Kyma cluster
type: Details
---

Restoring a Kyma cluster requires installing Velero. Velero CLI can be downloaded and used to install Velero server. Then, using the client, instruct Velero to start the restore process. Restore first the CRDs, Services, and Endpoints followed by the rest of resources.

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

> **NOTE:** Because of [this bug](https://github.com/heptio/velero/issues/1633) in Velero, sometimes the Custom Resources are not properly restored. In this case, you can run the second restore command again and check if the Custom Resources are restored. For example, this should print several VirtualService Custom Resources:

```
kubectl get virtualservices --all-namespaces
```
