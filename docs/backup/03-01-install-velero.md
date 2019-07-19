---
title: Install velero
type: Details
---
In order to be able to backup and restore your Kyma cluster, [Velero](https://github.com/heptio/velero/) needs to be configured and installed.

## Velero setup

To sucessfully set up velero, you need to provide a supported storage location in the cloud provider of your choosing and credentials to access it.
Please follow the instructions below to setup velero.
1. Enable the installation of the velero components on the kyma installer. To do that follow [this guide](/root/kyma/#configuration-custom-component-installation).<br/>
    The components to add are:
    ```yaml
    - name: "velero-essentials"
      namespace: "kyma-backup"
    - name: "velero"
      namespace: "kyma-backup"
    ```

2. Create an override secret containing the velero [required parameters](/components/backup/#configuration-configuration) for the provider of your choosing **encoded in base 64**.<br/>
    See the examples below for reference (_values are in plain text for illustrative purposes, remember to set them as base64 encoded strings_):

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
    >**NOTE:** For more information on configuring and installing velero in GCP visit https://velero.io/docs/v1.0.0/gcp-config/
    </details>
    <details>
    <summary>
    Azure
    </summary>

    Coming soon...

    >**NOTE:** For more information on configuring and installing velero in Azure visit https://velero.io/docs/v1.0.0/azure-config/
    
    </details>
    <details>
    <summary>
    AWS
    </summary>

    AWS is currently not officially supported.

    >**NOTE:** For more information on configuring and installing velero in AWS visit https://velero.io/docs/v1.0.0/aws-config/

    </details>
    </div>

    >**NOTE:** For more information on Kyma overrides visit the [Installation Overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation) section.

2. Run the kyma installation providing the velero overrides:
      <div tabs>
      <details>
      <summary>
      Local installation
      </summary>

      ```bash
      kyma install -o velero-overrides.yaml
      ```
      
      </details>
      <details>
      <summary>
      Cluster Installation
      </summary>
      
      1. Apply the overrides to your cluster:
          ```bash
          kubectl apply -f velero-overrides.yaml
          ```
      2. Install Kyma following the [installation guide](/root/kyma/#installation-installation) or update kyma if it is already installed in your cluster following the [update guide](/root/kyma/#installation-update-kyma).
      
      </details>
      </div>
