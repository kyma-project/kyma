---
title: Restore a Kyma cluster
type: Tutorial
---

Follow this tutorial to restore a backed up Kyma cluster. Start with restoring CRDs, services, and endpoints, then restore other resources.

## Prerequisites

To use the restore functionality, download and install the [Velero CLI](https://github.com/heptio/velero/releases/tag/v1.2.0) based on the `appVersion` in [Chart.yaml](https://github.com/kyma-project/kyma/tree/master/resources/backup/Chart.yaml).

## Steps

Follow these steps to restore resources:

1. Install the Velero server. Use the same bucket as for backups:

    <div tabs name="override-configuration">
      <details>
      <summary label="google-cloud-platform">
      Google Cloud Platform
      </summary>

      ```bash
      velero install \
          --provider gcp \
          --bucket {BUCKET} \
          --secret-file {CREDENTIALS_FILE} \
          --plugins velero/velero-plugin-for-gcp:v1.0.0,eu.gcr.io/kyma-project/backup-plugins:c08e6274 \
          --restore-only \
          --wait
      ```

      >**NOTE:** For details on configuring and installing Velero on GCP, see [this](https://github.com/vmware-tanzu/velero-plugin-for-gcp) repository.

      </details>
      <details>
      <summary label="azure">
      Azure
      </summary>

      ```bash
      velero install \
          --provider azure \
          --bucket {BUCKET} \
          --secret-file {CREDENTIALS_FILE} \
          --plugins velero/velero-plugin-for-microsoft-azure:v1.0.0,eu.gcr.io/kyma-project/backup-plugins:c08e6274 \
          --backup-location-config resourceGroup={AZURE_RESOURCE_GROUP},storageAccount={AZURE_STORAGE_ACCOUNT} \
          --snapshot-location-config apiTimeout={API_TIMEOUT},resourceGroup={AZURE_RESOURCE_GROUP} \
          --restore-only \
          --wait
      ```

      >**NOTE:** For details on configuring and installing Velero in Azure, see [this](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure) repository.

      >**CAUTION:** If you are using AKS, set the **AZURE_RESOURCE_GROUP** to the name of the auto-generated resource group created when you provision your cluster on Azure since this resource group contains your cluster's virtual machines/disks.

      </details>
    </div>

2. List available backups:

    ```bash
    velero get backups
    ```

3. Restore Kyma CRDs, services, and endpoints:

    ```bash
    velero restore create --from-backup <BACKUP_NAME> --include-resources customresourcedefinitions.apiextensions.k8s.io,services,endpoints --wait
    ```

4. Patch Velero deployment:

    ```bash
    kubectl patch deployment -n velero velero -p '
        {  
          "spec": {
            "template": {
              "spec": {
                "containers": [
                  {
                    "args": [
                      "server",
                      "--restore-resource-priorities=namespaces,persistentvolumes,persistentvolumeclaims,secrets,configmaps,serviceaccounts,limitranges,pods,clusterbuckets.rafter.kyma-project.io,buckets.rafter.kyma-project.io,  clusterassets.rafter.kyma-project.io,assets.rafter.kyma-project.io"
                    ],
                    "name": "velero"
                  }
                ]
              }
            }
          }
        }
        '
    ```

5. Restore the rest of Kyma resources:

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

6. Once the restore succeeds, remove the `velero` Namespace:

    ```bash
    kubectl delete ns velero
    ```
