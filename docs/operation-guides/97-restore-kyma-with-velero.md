---
title: Backup resources using Velero
---

Use Velero to back up individual applications running on Kyma and to restore Kubernetes resources and volumes, so that you can restore them on a different cluster.

> **NOTE:** Be aware that a full backup of a Kyma cluster is not supported. Start with the existing Kyma installation and restore specific resources individually.

## Prerequisites

Download and install [Velero CLI](https://github.com/vmware-tanzu/velero/releases).

## Steps

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

      >**NOTE:** For details on configuring and installing Velero on GCP, read more about the [Velero plugin for GCP](https://github.com/vmware-tanzu/velero-plugin-for-gcp/blob/master/README.md).

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

      >**NOTE:** For details on configuring and installing Velero on Azure, read more about the [Velero plugin for Azure](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure/blob/master/README.md).

      >**CAUTION:** If you use AKS, set the **AZURE_RESOURCE_GROUP** to the name of the auto-generated resource group. This resource group is created when you provision your cluster on Azure. It contains your cluster's virtual machines/disks.

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
