# Asset Store

## Overview

This chart consists of the following items:
* Asset Store Controller Manager
* Minio

## Change Minio to the Google Cloud Storage (GCS) Gateway mode

To configure Minio as a Gateway mode, you need a Secret with a service account that has the **Storage Admin** role permissions.

### Create a Secret

1. Open the [service accounts page](https://console.cloud.google.com/iam-admin/serviceaccounts).
2. Select one of the projects or create a new one. Note down the project ID as you must use it to update the Asset Store deployment.
3. Click **Create service account**, name your account, and click **Create**.
4. Set the **Storage Admin** role.
5. Click **Create key** and choose `JSON` as a key type.
6. Save the `JSON` file.
7. Create a Secret from the `JSON` file by running this command:
    ```bahs
    kubectl create secret generic assetstore-gcs-credentials --from-file=service-account.json={filename} --namespace kyma-system
    ```

### Update the Asset Store deployment

To update the Asset Store deployment, run this command:

```bash
helm upgrade assetstore resources/assetstore --namespace kyma-system --wait=true --reuse-values --set minio.persistence.enabled=false --set minio.gcsgateway.enabled=true --set minio.gcsgateway.replicas=1 --set minio.gcsgateway.gcsKeySecret=assetstore-gcs-credentials --set minio.gcsgateway.projectId={gcp-project} --set minio.defaultBucket.enabled=false
```

>**NOTE:** This is an alpha version of the Asset Store. In this version, the GCS `content` bucket is not available in the Minio Gateway mode.
