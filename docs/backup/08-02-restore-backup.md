---
title: Restore a Kyma cluster
type: Tutorial
---

Follow this tutorial to restore a backed up Kyma cluster. Start with restoring CRDs, services, and endpoints, then restore other resources.

## Prerequisites

To use the restore functionality, download and install the [Velero CLI](https://github.com/heptio/velero/releases/tag/v1.2.0).

## Steps

Follow these steps to restore resources:

1. Install the Velero server. Use the same bucket as for backups:

    ```bash
    velero install \
        --bucket {BUCKET} \
        --provider {CLOUD_PROVIDER} \
        --secret-file {CREDENTIALS_FILE} \
        --plugins velero/velero-plugin-for-gcp:v1.0.0,eu.gcr.io/kyma-project/backup-plugins:c08e6274 \
        --restore-only \
        --wait
    ```

    >**NOTE**: Check out this [guide](https://velero.io/docs/v1.2.0/customize-installation/) to correctly fill the parameters of this command corresponding to the cloud provider in use.

2. List available backups:

    ```bash
    velero get backups
    ```

3. Restore Kyma CRDs, services, and endpoints:

    ```bash
    velero restore create --from-backup <BACKUP_NAME> --include-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --wait
    ```

4. Restore the rest of Kyma resources:

    ```bash
    velero restore create --from-backup <BACKUP_NAME> --exclude-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --restore-volumes --wait
    ```

    Once the status of the restore is `COMPLETED`, perform a Kyma health check by verifying the Pods:

    ```bash
    kubectl get pods --all-namespaces
    ```

    Even if the restore process is complete, it may take some time for the resources to become available again.

    > **NOTE:** Because of [this issue](https://github.com/vmware-tanzu/velero/issues/964) in Velero, custom resources may not be properly restored. In this case, run the second restore command again and check if the custom resources are restored. For example, run the following command to print several VirtualService custom resources:
    >```bash
    > kubectl get virtualservices --all-namespaces
    > ```

5. Once the restore succeeds, remove the `velero` namespace:

    ```bash
    kubectl delete ns velero
    ```
