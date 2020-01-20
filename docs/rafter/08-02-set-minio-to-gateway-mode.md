---
title: Set MinIO to Gateway mode
type: Tutorials
---

By default, you install Kyma with Rafter in MinIO stand-alone mode. This tutorial shows how to set MinIO to Gateway mode on different cloud providers using an [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation).

>**CAUTION:** The authentication and authorization measures required to edit the assets in the public cloud storage may differ from those used in Rafter. That's why we recommend using separate subscriptions for Minio Gateway to ensure that you only have access to data created by Rafter, and to avoid compromising other public data.

## Prerequisites

<div tabs name="prerequisites" group="gateway-mode">
  <details>
  <summary label="google-cloud-storage">
  Google Cloud Storage
  </summary>

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [Google Cloud Platform (GCP)](https://cloud.google.com) project

  </details>
  <details>
  <summary label="azure-blob-storage">
  Azure Blob Storage
  </summary>

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure)
- [Microsoft Azure](http://azure.com) subscription

  </details>
  <details>
  <summary label="aws-s3">
  AWS S3
  </summary>

>**CAUTION:** AWS S3 Gateway mode was only tested manually on Kyma 1.6. Currently, there is no automated pipeline for it in Kyma.

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Amazon Web Services (AWS)](https://aws.amazon.com) account

  </details>
  <details>
  <summary label="alibaba-cloud-oss">
  Alibaba Cloud OSS
  </summary>

>**CAUTION:** Alibaba Cloud OSS Gateway mode was only tested manually on Kyma 1.6. Currently, there is no automated pipeline for it in Kyma.

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Alibaba Cloud](https://alibabacloud.com) account

  </details>
</div>


## Steps

You can set MinIO to the given Gateway mode both during and after Kyma installation. In both cases, you need to create and configure an access key for your cloud provider account, apply a Secret and a ConfigMap with an override to a cluster or Minikube, and trigger the Kyma installation process.

>**CAUTION:** Buckets created in MinIO without using Bucket CRs are not recreated or migrated while switching to MinIO Gateway mode.

### Create required cloud resources

<div tabs name="create-required-cloud-resources" group="gateway-mode">
  <details>
  <summary label="google-cloud-storage">
  Google Cloud Storage
  </summary>

Create a Google service account that has a private key and the **Storage Admin** role permissions. Follow these steps:

1. Run the `export {VARIABLE}={value}` command to set up the following environment variables, where:

    - **SA_NAME** is the name of the service account.
    - **SA_DISPLAY_NAME** is the display name of the service account.
    - **PROJECT** is the GCP project ID.
    - **SECRET_FILE** is the path to the private key.
    - **ROLE** is the **Storage Admin** role bound to the service account.

    Example:

    ```bash
    export SA_NAME=my-service-account
    export SA_DISPLAY_NAME=service-account
    export PROJECT=test-project-012345
    export SECRET_FILE=my-private-key-path
    export ROLE=roles/storage.admin
    ```

2. When you communicate with Google Cloud for the first time, set the context for your Google Cloud project. Run this command:

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

6. Export the private key as an environment variable:

    ```bash
    export GCS_KEY_JSON=$(< "$SECRET_FILE" base64 | tr -d '\n')
    ```

  </details>
  <details>
  <summary label="azure-blob-storage">
  Azure Blob Storage
  </summary>

Create an Azure resource group and a storage account. Follow these steps:

1. Run the `export {VARIABLE}={value}` command to set up the following environment variables, where:

    - **AZ_ACCOUNT_NAME** is the name of the storage account.
    - **AZ_RESOURCE_GROUP** is the name of the resource group.
    - **AZ_RESOURCE_GROUP_LOCATION** is the location of the resource group.
    - **AZ_SUBSCRIPTION** is the ID of the Azure subscription.

    Example:

    ```bash
    export AZ_ACCOUNT_NAME=accountname
    export AZ_RESOURCE_GROUP=my-resource-group
    export AZ_RESOURCE_GROUP_LOCATION=westeurope
    export AZ_SUBSCRIPTION=123456-123456-123456-1234567
    ```

2. When you communicate with Microsoft Azure for the first time, log into your Azure account. Run this command:

    ```bash
    az login
    ```

3. Create a resource group. Run:

    ```bash
    az group create --name ${AZ_RESOURCE_GROUP} --location ${AZ_RESOURCE_GROUP_LOCATION} --subscription ${AZ_SUBSCRIPTION}
    ```

4. Create a storage account. Run:

    ```bash
    az storage account create --name ${AZ_ACCOUNT_NAME} --resource-group ${AZ_RESOURCE_GROUP} --subscription ${AZ_SUBSCRIPTION}
    ```

5. Export the access key as an environment variable:

    ```bash
    export AZ_ACCOUNT_KEY=$(az storage account keys list --account-name "${AZ_ACCOUNT_NAME}" --resource-group "${AZ_RESOURCE_GROUP}" --query "[?keyName=='key1'].value" --output tsv | base64)
    ```

  </details>
  <details>
  <summary label="aws-s3">
  AWS S3
  </summary>

Create an AWS access key for an IAM user. Follow these steps:

1. Access the [AWS Identity and Access Management console](https://console.aws.amazon.com/iam/).
2. In the left navigation panel, select **Users**.
3. Choose the user whose access keys you want to create, and select the **Security credentials** tab.
4. In the **Access keys** section, select **Create access key**.
5. Export the access and secret keys as environment variables:

    ```bash
    export AWS_ACCESS_KEY={your-access-ID}
    export AWS_SECRET_KEY={your-secret-key}
    ```

  </details>
  <details>
  <summary label="alibaba-cloud-oss">
  Alibaba Cloud OSS
  </summary>

Create an Alibaba Cloud access key for a user. Follow these steps:

1. Access the [Resource Access Management console](https://ram.console.aliyun.com).
2. In the left navigation panel, select **User**.
3. Choose the user whose access keys you want to create.
4. Click **Create AccessKey** in the **User AccessKey** section.
5. Export the access and secret keys as environment variables:

    ```bash
    export ALIBABA_ACCESS_KEY={your-access-ID}
    export ALIBABA_SECRET_KEY={your-secret-key}
    ```

  </details>
</div>

### Configure MinIO Gateway mode

<div tabs name="configure-minio-gateway-mode" group="gateway-mode">
  <details>
  <summary label="google-cloud-storage">
  Google Cloud Storage
  </summary>


Apply the following Secret and ConfigMap with an override to a cluster or Minikube. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: rafter-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: rafter
    kyma-project.io/installation: ""
type: Opaque
data:
  controller-manager.minio.gcsgateway.gcsKeyJson: "$GCS_KEY_JSON"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: rafter-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: rafter
    kyma-project.io/installation: ""
data:
  controller-manager.minio.persistence.enabled: "false"
  controller-manager.minio.gcsgateway.enabled: "true"
  controller-manager.minio.gcsgateway.projectId: "$PROJECT"
  controller-manager.minio.DeploymentUpdate.type: RollingUpdate
  controller-manager.minio.DeploymentUpdate.maxSurge: "0"
  controller-manager.minio.DeploymentUpdate.maxUnavailable: "50%"
EOF
```

  </details>
  <details>
  <summary label="azure-blob-storage">
  Azure Blob Storage
  </summary>

Apply the following Secret and ConfigMap with an override to a cluster or Minikube. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: rafter-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: rafter
    kyma-project.io/installation: ""
type: Opaque
data:
  controller-manager.minio.accessKey: "$(echo "${AZ_ACCOUNT_NAME}" | base64)"
  controller-manager.minio.secretKey: "${AZ_ACCOUNT_KEY}"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: rafter-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: rafter
    kyma-project.io/installation: ""
data:
  controller-manager.minio.persistence.enabled: "false"
  controller-manager.minio.azuregateway.enabled: "true"
  controller-manager.minio.DeploymentUpdate.type: RollingUpdate
  controller-manager.minio.DeploymentUpdate.maxSurge: "0"
  controller-manager.minio.DeploymentUpdate.maxUnavailable: "50%"
EOF
```

  </details>
  <details>
  <summary label="aws-s3">
  AWS S3
  </summary>

1. Export an AWS S3 service [endpoint](https://docs.aws.amazon.com/general/latest/gr/rande.html) as an environment variable:

    ```bash
    export AWS_SERVICE_ENDPOINT=https://{endpoint-address}
    ```

2. Apply the following Secret and ConfigMap with an override to a cluster or Minikube. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: rafter-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: rafter
    kyma-project.io/installation: ""
type: Opaque
data:
  controller-manager.minio.accessKey: "$(echo "${AWS_ACCESS_KEY}" | base64)"
  controller-manager.minio.secretKey: "$(echo "${AWS_SECRET_KEY}" | base64)"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: rafter-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: rafter
    kyma-project.io/installation: ""
data:
  controller-manager.minio.persistence.enabled: "false"
  controller-manager.minio.s3gateway.enabled: "true"
  controller-manager.minio.s3gateway.serviceEndpoint: "${AWS_SERVICE_ENDPOINT}"
  controller-manager.minio.DeploymentUpdate.type: RollingUpdate
  controller-manager.minio.DeploymentUpdate.maxSurge: "0"
  controller-manager.minio.DeploymentUpdate.maxUnavailable: "50%"
EOF
```

  </details>
  <details>
  <summary label="alibaba-cloud-oss">
  Alibaba Cloud OSS
  </summary>

1. Export an Alibaba OSS service [endpoint](https://www.alibabacloud.com/help/doc-detail/31837.htm) as an environment variable:

    ```bash
    export ALIBABA_SERVICE_ENDPOINT=https://{endpoint-address}
    ```

2. Apply the following Secret and ConfigMap with an override to a cluster or Minikube. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: rafter-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: rafter
    kyma-project.io/installation: ""
type: Opaque
data:
  controller-manager.minio.accessKey: "$(echo "${ALIBABA_ACCESS_KEY}" | base64)"
  controller-manager.minio.secretKey: "$(echo "${ALIBABA_SECRET_KEY}" | base64)"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: rafter-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: rafter
    kyma-project.io/installation: ""
data:
  controller-manager.minio.persistence.enabled: "false"
  controller-manager.minio.ossgateway.enabled: "true"
  controller-manager.minio.ossgateway.endpointURL: "${ALIBABA_SERVICE_ENDPOINT}"
  controller-manager.minio.DeploymentUpdate.type: RollingUpdate
  controller-manager.minio.DeploymentUpdate.maxSurge: "0"
  controller-manager.minio.DeploymentUpdate.maxUnavailable: "50%"
EOF
```

  </details>
</div>

>**CAUTION:** When you install Kyma locally from sources, you need to manually add the ConfigMap and the Secret to the `installer-config-local.yaml.tpl` template located under the `installation/resources` subfolder before you run the installation script.

### Trigger installation

Trigger Kyma installation or update by labeling the Installation custom resource. Run:

```bash
kubectl label installation/kyma-installation action=install
```
