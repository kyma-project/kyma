---
title: Taking backup using Velero
type: Tutorials
---

This tutorial shows how to use Velero to perform a partial restore of individual applications running on Kyma. Follow the guidelines to back up your Kubernetes resources and volumes so that you can restore them on a different cluster.

> **NOTE:**  Be aware that a full restore of a Kyma cluster is not supported. You should start with an existing Kyma installation and restore specific resources individually.


## Prerequisites

Download and install the [Velero CLI](https://github.com/vmware-tanzu/velero/releases).

## Steps

Follow these steps to install Velero and back up your Kyma cluster.

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

      >**NOTE:** For details on configuring and installing Velero on GCP, see [this](https://github.com/vmware-tanzu/velero-plugin-for-gcp/blob/master/README.md) document.

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

      >**NOTE:** For details on configuring and installing Velero on Azure, see [this](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure/blob/master/README.md) document.

      >**CAUTION:** If you are using AKS, set the **AZURE_RESOURCE_GROUP** to the name of the auto-generated resource group created when you provision your cluster on Azure since this resource group contains your cluster's virtual machines/disks.

      </details>
    </div>

2. Create a backup of all the resources on the cluster:

    ```bash
    velero backup create {NAME} --wait
    ```

3. Once the backup succeeds, remove the `velero` Namespace:

    ```bash
    kubectl delete ns velero
    ```
