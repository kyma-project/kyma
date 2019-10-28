---
title: Set MinIO to the Azure Blob Storage Gateway mode
type: Tutorials
---

By default, you install Kyma with the Asset Store in MinIO stand-alone mode. This tutorial shows how to set MinIO to the Azure Blob Storage Gateway mode using an [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation).

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure)
- [Microsoft Azure](http://azure.com) subscription

## Steps

You can set MinIO to the Azure Blob Storage Gateway mode both during and after Kyma installation. In both cases, you need to create and configure an Azure storage account, apply a Secret and ConfigMap with an override onto a cluster or Minikube, and trigger the Kyma installation process.

>**CAUTION:** Buckets created in MinIO without using Bucket CRs are not recreated or migrated while switching to the MinIO Gateway mode.

### Set up Azure Blob Storage resources

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

### Configure MinIO Gateway mode

Apply the following Secret and ConfigMap with an override onto a cluster or Minikube. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: asset-store-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: assetstore
    kyma-project.io/installation: ""
type: Opaque
data:
  minio.accessKey: "$(echo "${AZ_ACCOUNT_NAME}" | base64)"
  minio.secretKey: "${AZ_ACCOUNT_KEY}"
---
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
  minio.azuregateway.enabled: "true"
  minio.DeploymentUpdate.type: RollingUpdate
  minio.DeploymentUpdate.maxSurge: "0"
  minio.DeploymentUpdate.maxUnavailable: "50%"
EOF
```

>**CAUTION:** When you install Kyma locally from sources, you need to manually add the ConfigMap and the Secret to the `installer-config-local.yaml.tpl` template located under the `installation/resources` subfolder before you run the installation script.

### Trigger installation

Trigger Kyma installation or update by labeling the Installation custom resource. Run:

```bash
kubectl label installation/kyma-installation action=install
```
