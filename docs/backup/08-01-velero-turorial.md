---
title: Taking backup using Velero
type: Tutorials
---

Follow this tutorial to install Velero and take a backup from Kyma cluster.

## Prerequisites

Download and install the [Velero CLI](https://github.com/heptio/velero/releases/tag/v1.3.2).

## Steps

Follow these steps to install Velero and take a backup:

1. Install the Velero server.

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
          --plugins velero/velero-plugin-for-gcp:v1.0.0 \
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
          --plugins velero/velero-plugin-for-microsoft-azure:v1.0.0 \
          --backup-location-config resourceGroup={AZURE_RESOURCE_GROUP},storageAccount={AZURE_STORAGE_ACCOUNT} \
          --snapshot-location-config apiTimeout={API_TIMEOUT},resourceGroup={AZURE_RESOURCE_GROUP} \
          --restore-only \
          --wait
      ```

      >**NOTE:** For details on configuring and installing Velero in Azure, see [this](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure) repository.

      >**CAUTION:** If you are using AKS, set the **AZURE_RESOURCE_GROUP** to the name of the auto-generated resource group created when you provision your cluster on Azure since this resource group contains your cluster's virtual machines/disks.

      </details>
    </div>

2. Take a backup of all the resources on the cluster:

    ```bash
    velero backup create {NAME} --wait
    ```

3. Once the backup succeeds, remove the `velero` Namespace:

    ```bash
    kubectl delete ns velero
    ```
