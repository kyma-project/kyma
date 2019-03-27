---
title: Set Minio to the Google Cloud Storage (GCS) Gateway mode
type: Tutorials
---

This tutorial shows how to set Minio to the Google Cloud Storage (GCS) Gateway mode.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl)
- [helm](https://github.com/helm/helm#install)
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [Google Cloud Platform (GCP)](https://cloud.google.com) project

## Steps

Follow these guidelines to create a Google service account and a Secret, and to update the Asset Store deployment.

### Create a Google service accounts

Run the `export {VARIABLE}={value}` command to set up the following environment variables, where:

- **SA_NAME** is the name of the service account.
- **SA_DISPLAY_NAME** is the display name of the service account.
- **PROJECT** is the GCP project ID.
- **SECRET_FILE** is the path to the private key.
- **ROLE** is the **Storage Admin** role bound to the service account.

Example:
```
export SA_NAME=my-service-account
export SA_DISPLAY_NAME=service-account
export PROJECT=service-account-012345
export SECRET_FILE=my-private-key-path
export ROLE=roles/storage.admin
```

### Create a Secret

When you communicate with Google Cloud for the first time, set context to your Google Cloud project. Run this command:
```bash
gcloud config set project $PROJECT
```

To set Minio to a Gateway mode, you need a Secret with a service account that has the **Storage Admin** role permissions. Follow these steps:

1. Create a service account. Run:
    ```bash
    gcloud iam service-accounts create $SA_NAME --display-name $SA_DISPLAY_NAME
    ```
2. Add a policy binding for the **Storage Admin** role to the service account. Run:
    ```bash
    gcloud projects add-iam-policy-binding $PROJECT --member=serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com --role=$ROLE
    ```
3. Create a private key for the service account:
    ```bash
    gcloud iam service-accounts keys create $SECRET_FILE --iam-account=$SA_NAME@$PROJECT.iam.gserviceaccount.com
    ```
4. Create a Secret:
    ```bash
    kubectl create secret generic assetstore-gcs-credentials --from-file=service-account.json=$SECRET_FILE --namespace kyma-system
    ```

### Update the Asset Store deployment

Go to the `kyma` directory and update the Asset Store deployment by running this command:

```bash
helm upgrade assetstore resources/assetstore --namespace kyma-system --wait=true --reuse-values --set minio.persistence.enabled=false --set minio.gcsgateway.enabled=true --set minio.gcsgateway.replicas=1 --set minio.gcsgateway.gcsKeySecret=assetstore-gcs-credentials --set minio.gcsgateway.projectId=$PROJECT --set minio.defaultBucket.enabled=false
```
