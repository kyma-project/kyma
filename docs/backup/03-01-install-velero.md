---
title: Install velero into Kyma
type: Details
---
In order to be able to backup and restore your Kyma cluster, [Velero](https://github.com/heptio/velero/) needs to be configured and installed.

## Velero configuration

To sucessfully configure velero, you need to provide a supported storage location and credentials to access it. Please follow the instructions below for the provider of your choosing:
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
      ```bash
        kyma install -o velero-overrides.yaml
      ```

  >**NOTE:** For more information visit https://velero.io/docs/v1.0.0/gcp-config/

  </details>
  <details>
  <summary>
  Azure
  </summary>

  Coming soon...

  >**NOTE:** For more information visit https://velero.io/docs/v1.0.0/azure-config/
  
  </details>
  <details>
  <summary>
  AWS
  </summary>

  AWS is currently not officially supported.

  >**NOTE:** For more information visit https://velero.io/docs/v1.0.0/aws-config/

  </details>
  </div>