---
title: Set Minio to the Google Cloud Storage Gateway mode
type: Tutorials
---

By default, you install Kyma with the Asset Store in Minio stand-alone mode. This tutorial shows how to set Minio to the Google Cloud Storage (GCS) Gateway mode using an [override](root/kyma/#tutorials-helm-overrides-for-kyma-installation).

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [Google Cloud Platform (GCP)](https://cloud.google.com) project

## Steps

You can set Minio to the GCS Gateway mode during Kyma installation, or switch to the GCS Gateway mode after Kyma installation. In both cases you need to create a Google service account that has a private key and the **Storage Admin** role permissions, apply a ConfigMap with an override onto a cluster or Minikube, and trigger Kyma installation. Follow these steps:

1. Run the `export {VARIABLE}={value}` command to set up the following environment variables, where:

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
2. When you communicate with Google Cloud for the first time, set context to your Google Cloud project. Run this command:
    ```bash
    gcloud config set project $PROJECT
    ```
3. Create a service account. Run:
    ```bash
    gcloud iam service-accounts create $SA_NAME --display-name $SA_DISPLAY_NAME
    ```
4. Add a policy binding for the **Storage Admin** role to the service account. Run:
    ```bash
    gcloud projects add-iam-policy-binding $PROJECT --member=serviceAccount:$SA_NAME@$PROJECT.iam.gserviceaccount.com --role=$ROLE
    ```
5. Create a private key for the service account:
    ```bash
    gcloud iam service-accounts keys create $SECRET_FILE --iam-account=$SA_NAME@$PROJECT.iam.gserviceaccount.com
    ```
6. Export the private key as the environment variable:
    ```bash
    export GCS_KEY_JSON=$(< "${SECRET_FILE}" base64 | tr -d '\n')
    ```
7. Apply the following ConfigMap with an override onto a cluster or Minikube. Run:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ConfigMap
    metadata:
    name: asset-store-overrides
    namespace: kyma-installer
    labels:
        installer: overrides
        component: assetstore
        kyma-project.io/installation: ""
    data:
    minio.persistence.enabled: "false"
    minio.gcsgateway.enabled: "true"
    minio.defaultBucket.enabled: "false"
    minio.gcsgateway.projectId: "${PROJECT}"
    minio.gcsgateway.gcsKeyJson: "${GCS_KEY_JSON}"
    minio.externalEndpoint: "https://storage.googleapis.com"
    EOF
    ```
8. Lable the `kyma-installtion` custom resource by running `kubectl label installation/kyma-installation action=install` to trigger Kyma installation.


>**CAUTION** When you install Kyma locally from sources, you need to manually add the ConfigMap to the `installer-config-local.yaml.tpl` template located under the `installation/resources` subfolder before you run the installation script.