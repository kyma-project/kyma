---
title: Install Velero
type: Installation
---

Install and configure [Velero](https://github.com/heptio/velero/) to back up and restore your Kyma cluster.

>**NOTE**: To successfully set up Velero, define a supported storage location and credentials to access it. Currently, you can install Velero on GCP and Azure. AWS is not supported.

Follow the instructions to set up Velero: 

1. Override the default backup configuration provided by the `backup` and `backup-init` components by creating a Secret containing the [required parameters](/components/backup/#configuration-configuration) for a chosen provider. 

    See examples of such Secrets:

    >**NOTE**: The values are provided in plain text for illustrative purposes only. Remember to set them as base64-encoded strings. For details on Kyma overrides, see [this](/root/kyma/#configuration-helm-overrides-for-kyma-installation) document.

    <div tabs name="override-configuration">
      <details>
      <summary label="google-cloud-platform">
      Google Cloud Platform
      </summary>
        
      ```yaml
      apiVersion: v1
      kind: Secret
      metadata:
        name: velero-credentials-overrides
        namespace: kyma-installer
        labels:
          kyma-project.io/installation: ""
          installer: overrides
          component: backup
      type: Opaque
      stringData:
        configuration.provider: "gcp"
        configuration.volumeSnapshotLocation.name: "gcp"
        configuration.volumeSnapshotLocation.bucket: "my-gcp-bucket"
        configuration.backupStorageLocation.name: "gcp"
        configuration.backupStorageLocation.bucket: "my-gcp-bucket"
        credentials.secretContents.cloud: |
                    {
                        "type": "service_account",
                        "project_id": "my-project",
                        "private_key_id": "KEY_UUID",
                        "private_key": "-----BEGIN PRIVATE KEY-----\nPRIVATE_KEY_CONTENTS\n-----END PRIVATE KEY-----\n",
                        "client_email": "sample@fake.iam.gserviceaccount.com",
                        "client_id": "MY_CLIENT_ID",
                        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
                        "token_uri": "https://oauth2.googleapis.com/token",
                        "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
                        "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/sample%40fake.iam.gserviceaccount.com"
                    }
      ```
    
      >**NOTE:** For details on configuring and installing Velero on GCP, see [this](https://github.com/vmware-tanzu/velero-plugin-for-gcp) repo.
      </details>
      <details>
      <summary label="azure">
      Azure
      </summary>

      ```yaml
      apiVersion: v1
      kind: Secret
      metadata:
        name: velero-credentials-overrides
        namespace: kyma-installer
        labels:
          kyma-project.io/installation: ""
          installer: overrides
          component: backup
      type: Opaque
      stringData:
        configuration.provider: "azure"
        configuration.volumeSnapshotLocation.name: "azure"
        configuration.volumeSnapshotLocation.bucket: "my-storage-container"
        configuration.volumeSnapshotLocation.config.apitimeout: "3m0s"
        configuration.backupStorageLocation.name: "azure"
        configuration.backupStorageLocation.bucket: "my-storage-container"
        configuration.backupStorageLocation.config.resourceGroup: "my-resource-group"
        configuration.backupStorageLocation.config.storageAccount: "my-storage-account"
        credentials.secretContents.cloud: |
                        AZURE_SUBSCRIPTION_ID=my-subscription-ID
                        AZURE_TENANT_ID=my-tenant-ID
                        AZURE_CLIENT_ID=my-client-ID
                        AZURE_CLIENT_SECRET=my-client-secret
                        AZURE_RESOURCE_GROUP=my-resource-group
      ```

      >**NOTE:** For details on configuring and installing Velero in Azure, see [this](https://github.com/vmware-tanzu/velero-plugin-for-aws) repo.
        
      </details>
    </div>

2. Run the Kyma installation with Velero overrides:

    <div tabs name="run-velero">
      <details>
      <summary label="local-installation">
      Local installation
      </summary>

      To apply overrides to your local installation, run:

      ```bash
      kyma install -o {overrides_file_path}
      ```
      
      </details>
      <details>
      <summary label="cluster-installation">
      Cluster installation
      </summary>
      
      1. Apply the overrides to your cluster:

        ```bash
        kubectl apply -f {overrides_file_path}
        ```

      2. [Install](/root/kyma/#installation-installation) Kyma or [update](/root/kyma/#installation-update-kyma) it if it is already installed on your cluster.
      
      </details>
    </div>
