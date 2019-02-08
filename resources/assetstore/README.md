# Asset Store

## Overview

The Service Catalog Add-ons provide Kyma add-ons to the [Service Catalog](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/README.md).

These chart consist of the following items:
* Asset Store Controller Manager
* Minio

## Change Minio to GCS Gateway mode

To configure Minio as a Gateway mode you will need a Secret with Service Account that has `Storage Admin` role.

### Create a Secret

1. Navigate to the [Service Accounts page](https://console.cloud.google.com/iam-admin/serviceaccounts).
2. Select a project or create a new project. Note the project ID.
3. Click **Create service account**, name your account, and click **Create**.
4. Set the `Storage Admin` role.
5. Click **Create key** and choose `JSON` as a key type.
6. Save the `JSON` file.
7. Create a Secret from the JSON file by running this command:
    ```bahs
    kubectl create secret generic assetstore-gcs-credentials --from-file=service-account.json={filename} --namespace kyma-system
    ```

### Update Asset Store deployment

To update Asset Store deployment run this command:

```bash
helm upgrade assetstore resources/assetstore --namespace kyma-system --wait=true --reuse-values --set minio.persistence.enabled=false --set minio.gcsgateway.enabled=true --set minio.gcsgateway.replicas=1 --set minio.gcsgateway.gcsKeySecret=assetstore-gcs-credentials --set minio.gcsgateway.projectId={gcp-project} --set minio.defaultBucket.enabled=false
```

>**NOTE:** This is alpha version of Asset Store and after migration to GCS `content` bucket will be not available. Also buckets in GCS must by globally unique.