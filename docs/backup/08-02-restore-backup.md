---
title: Restore a Kyma cluster
type: Tutorial
---

Follow this tutorial to restore a backed up Kyma cluster. Start with restoring CRDs, services, and endpoints, then restore other resources.

## Prerequisites

To use the restore functionality, download and install the [Velero CLI](https://github.com/heptio/velero/releases/tag/v1.0.0).


## Steps

Follow these steps to restore resources: 

1. Install the Velero server. Use the same bucket as for backups:

    ```bash
    velero install --bucket <BUCKET> --provider <CLOUD_PROVIDER> --secret-file <CREDENTIALS_FILE> --restore-only --wait
    ```

    >**NOTE**: Check out this [guide](https://velero.io/docs/v1.0.0/install-overview/) to correctly fill the parameters of this command corresponding to the cloud provider in use.

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

    Once the status of the restore is `COMPLETED`, perform a Kyma health check by verifying the Pods:

    ```bash
    kubectl get pods --all-namespaces
    ```

    Even if the restore process is complete, it may take some time for the resources to become available again.

    > **NOTE:** Because of [this issue](https://github.com/heptio/velero/issues/1633) in Velero, custom resources may not be properly restored. In this case, run the second restore command again and check if the custom resources are restored. For example, run the following command to print several VirtualService custom resources:
    >```bash
    > kubectl get virtualservices --all-namespaces
    > ```

5. Once the restore succeeds, remove the `velero` namespace:

    ```bash
    kubectl delete ns velero
    ```

## Troubleshooting

### Pod stuck in `Init` phase

In case the Pod `service-catalog-addons-service-binding-usage-controller` gets stuck in `Init` phase, try deleting the Pod:

```bash
kubectl delete $(kubectl get pod -l app=service-catalog-addons-service-binding-usage-controller -n kyma-system -o name) -n kyma-system
```

### Different DNS and Gateway IP Address

This tutorial assumes that the DNS and the Public IP will stay the same as the backed up cluster. If they change in the new cluster, check and update the relavant fields on the `overrides` Secrets and ConfigMaps in `kyma-installer` namespace with the new values and re-run the installer to propagate them to all the components:

```bash
kubectl label installation/kyma-installation action=install
```
