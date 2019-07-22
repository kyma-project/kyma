---
title: Install Velero
type: Details
---
Install and configure [Velero](https://github.com/heptio/velero/) to back up and restore your Kyma cluster.

## Velero setup

To successfully set up Velero, provide a supported storage location and credentials to access it. 

>**NOTE**: Currently, you can install Velero on GCP and Azure. AWS is not supported.

Follow the instructions below:
1. Enable Velero components in the Kyma Installer configuration file. To do that follow [this guide](/root/kyma/#configuration-custom-component-installation).<br/>
    Add the following components:
    ```yaml
    - name: "velero-essentials"
      namespace: "kyma-backup"
    - name: "velero"
      namespace: "kyma-backup"
    ```

2. Create an override secret containing the Velero [required parameters](/components/backup/#configuration-configuration) for a chosen provider. Remember to base64-encode the parameters.<br/>
    See the installation examples:
    >**NOTE**: Values are provided in plain text only for illustrative purposes. Remember to set them as base64-encoded strings.

    <div tabs>
    <details>
    <summary>
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
        component: velero
    type: Opaque
    data:
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
    >**NOTE:** For details on configuring and installing Velero in GCP,  see [this](https://velero.io/docs/v1.0.0/gcp-config/) document.
    </details>
    <details>
    <summary>
    Azure
    </summary>

    Coming soon...

    >**NOTE:** For details on configuring and installing Velero in Azure,  see [this](https://velero.io/docs/v1.0.0/azure-config/) document.
    
    </details>
    </div>

    >**NOTE:** For details on Kyma overrides visit the [Installation Overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation) section.

2. Run the Kyma installation providing the Velero overrides:
      <div tabs>
      <details>
      <summary>
      Local installation
      </summary>

      ```bash
      kyma install -o {overrides_file_path}
      ```
      
      </details>
      <details>
      <summary>
      Cluster installation
      </summary>
      
      1. Apply the overrides to your cluster:
          ```bash
          kubectl apply -f {overrides_file_path}
          ```
      2. [Install](/root/kyma/#installation-installation) Kyma or [update](/root/kyma/#installation-update-kyma) Kyma if it is already installed in your cluster.
      
      </details>
      </div>
