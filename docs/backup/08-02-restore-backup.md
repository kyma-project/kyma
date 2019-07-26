---
title: Restore a Kyma cluster
type: Tutorial
---

Follow this tutorial to restore a backed up Kyma cluster. Restore the CRDs, services, and endpoints first, and then the rest of the resources.

## Prerequisites

To use the restore functionality, dowwnload and install the [Velero CLI](https://github.com/heptio/velero/releases/tag/v1.0.0).


## Steps

Follow these steps to restore the resources: 

1. Install the Velero server. Use the same bucket as for the backups:

    ```bash
    velero install --bucket <BUCKET> --provider <CLOUD_PROVIDER> --secret-file <CREDENTIALS_FILE> --restore-only --wait
    ```

2. List available backups:

    ```bash
    velero get backups
    ```

3. Restore Kyma CRDs, services, and endpoints:

    ```bash
    velero restore create --from-backup <BACKUP_NAME> --include-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --include-cluster-resources --wait
    ```

4. Restore the rest of Kyma resources:

    ```bash
    velero restore create --from-backup <BACKUP_NAME> --exclude-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --include-cluster-resources --restore-volumes --wait
    ```

    Once the status of the restore is `COMPLETED`, verify the health of Kyma by checking the Pods:

    ```bash
    kubectl get pods --all-namespaces
    ```

    Even if the restore process is complete, it may take some time for the resources to become available again.

    > **NOTE:** Because of [this issue](https://github.com/heptio/velero/issues/1633) in Velero, Custom Resources may not be properly restored. In this case, run the second restore command again and check if the Custom Resources are restored. For example, run the following command to print several VirtualService Custom Resources:

    ```bash
    kubectl get virtualservices --all-namespaces
    ```
