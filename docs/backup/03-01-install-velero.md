---
title: Install velero
type: Details
---
In order to be able to backup and restore your Kyma cluster, [Velero](https://github.com/heptio/velero/) needs to be configured and installed.

## Velero setup

To sucessfully set up velero, you need to provide a supported storage location and credentials to access it. Please follow the instructions below for the provider of your choosing:
  <div tabs>
  <details>
  <summary>
  Google Cloud Platform
  </summary>
  
  For GCP, you need to have a bucket in `Google Cloud Storage`, and a service account able to access it and store data (and its JSON key).

  1. Download the overrides file to your installation folder:
      - [Velero Overrides](./assets/velero-overrides.yaml)

  2. Run the following commands providing the necessary information to replace the placeholders on velero's configuration with your GCP bucket and credentials:
      ```bash
          # Set the base64 encoded provider to gcp
          sed -i.bak "s/__PROVIDER__/$(echo -n gcp | base64)/g" velero-overrides.yaml

          # Set the base64 encoded bucket name to the name of your bucket in GCS
          sed -i.bak "s/__BUCKET__/$(echo -n <bucket name> | base64)/g" velero-overrides.yaml
          
          # Set the base64 encoded credentials JSON providing your service account key file
          sed -i.bak "s/__CREDENTIALS__/$(base64 <credentials JSON file path>)/g" velero-overrides.yaml
      ```

  3. Run the kyma installation with the velero overrides:
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
      
      1. Install Kyma following the [installation guide](/root/kyma/#installation-installation)
      
      2. Apply the overrides
          ```bash
          kubectl apply -f velero-overrides.yaml
          ```
      3. Trigger the installer to apply the overrides
          ```bash
          kubectl label installation/kyma-installation action=install
          ```
      
      </details>
      </div>

  >**NOTE:** For more information on Kyma overrides visit the [Installation Overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation) section.

  >**NOTE:** For more information on configuring and installing velero in GCP visit https://velero.io/docs/v1.0.0/gcp-config/

  </details>
  <details>
  <summary>
  Azure
  </summary>

  Coming soon...

  >**NOTE:** For more information on Kyma overrides visit the [Installation Overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation) section.

  >**NOTE:** For more information on configuring and installing velero in Azure visit https://velero.io/docs/v1.0.0/azure-config/
  
  </details>
  <details>
  <summary>
  AWS
  </summary>

  AWS is currently not officially supported.

  >**NOTE:** For more information on Kyma overrides visit the [Installation Overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation) section.

  >**NOTE:** For more information on configuring and installing velero in AWS visit https://velero.io/docs/v1.0.0/aws-config/

  </details>
  </div>