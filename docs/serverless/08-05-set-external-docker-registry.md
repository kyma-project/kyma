---
title: Set external Docker Registry
type: Tutorials
---

By default, you install Kyma with Serverless with internal (running on cluster) Docker Registry. This tutorial shows how to set external Docker Registry from different cloud providers using an [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation).

## Prerequisites

<div tabs name="prerequisites" group="external-docker-registry">
  <details>
  <summary label="docker-hub">
  Docker Hub
  </summary>

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

  </details>
  <details>
  <summary label="gcr">
  GCR
  </summary>

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [Google Cloud Platform (GCP)](https://cloud.google.com) project

  </details>
  <details>
  <summary label="azure-cr">
  Azure CR
  </summary>

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure)
- [Microsoft Azure](http://azure.com) subscription

  </details>
</div>

## Steps

### Create required cloud resources

<div tabs name="create-required-cloud-resources" group="external-docker-registry">
  <details>
  <summary label="docker-hub">
  Docker Hub
  </summary>

1. Run the `export {VARIABLE}={value}` command to set up the following environment variables, where:

    - **USERNAME** is the name of account in Docker Hub.
    - **PASSWORD** is the password for account in Docker Hub.
    - **SERVER_ADDRESS** is the server address of Docker Hub. At the moment Kyma supports only `https://index.docker.io/v1/` server address.
    - **REGISTRY_ADDRESS** is the registry address in Docker Hub (usually this is an account name).

    Example:

    ```bash
    export USERNAME=kyma-rocks
    export PASSWORD=admin123
    export SERVER_ADDRESS=https://index.docker.io/v1/
    export REGISTRY_ADDRESS=kyma-rocks
    ```

  </details>
  <details>
  <summary label="gcr">
  GCR
  </summary>

Create a Google service account that has a private key and the **Storage Admin** role permissions. Follow these steps:

1. Run the `export {VARIABLE}={value}` command to set up the following environment variables, where:

    - **SA_NAME** is the name of the service account.
    - **SA_DISPLAY_NAME** is the display name of the service account.
    - **PROJECT** is the GCP project ID.
    - **SECRET_FILE** is the path to the private key.
    - **ROLE** is the **Storage Admin** role bound to the service account.
    - **SERVER_ADDRESS** is the server address of Docker Registry.

    Example:

    ```bash
    export SA_NAME=my-service-account
    export SA_DISPLAY_NAME=service-account
    export PROJECT=test-project-012345
    export SECRET_FILE=my-private-key-path
    export ROLE=roles/storage.admin
    export SERVER_ADDRESS=grc.io
    ```

2. When you communicate with Google Cloud for the first time, set the context for your Google Cloud project. Run this command:

    ```bash
    gcloud config set project ${PROJECT}
    ```

3. Create a service account. Run:

    ```bash
    gcloud iam service-accounts create ${SA_NAME} --display-name ${SA_DISPLAY_NAME}
    ```

4. Add a policy binding for the **Storage Admin** role to the service account. Run:

    ```bash
    gcloud projects add-iam-policy-binding ${PROJECT} --member=serviceAccount:${SA_NAME}@${PROJECT}.iam.gserviceaccount.com --role=${ROLE}
    ```

5. Create a private key for the service account:

    ```bash
    gcloud iam service-accounts keys create ${SECRET_FILE} --iam-account=${SA_NAME}@${PROJECT}.iam.gserviceaccount.com
    ```

6. Export the private key as an environment variable:

    ```bash
    export GCS_KEY_JSON=$(< "${SECRET_FILE}")
    ```

  </details>
  <summary label="azure-cr">
  Azure CR
  </summary>

Create an Azure Container Registry and a service principal. Follow these steps:

1. Run the `export {VARIABLE}={value}` command to set up the following environment variables, where:

    - **AZ_REGISTRY_NAME** is the name of an Azure Container Registry.
    - **AZ_RESOURCE_GROUP** is the name of the resource group.
    - **AZ_RESOURCE_GROUP_LOCATION** is the location of the resource group.
    - **AZ_SUBSCRIPTION_ID** is the ID of the Azure subscription.
    - **AZ_SERVICE_PRINCIPAL_NAME** is the name of the Azure service principal.
    - **ROLE** is the **acrpush** role bound to the service principal.
    - **SERVER_ADDRESS** is the server address of Docker Registry.

    Example:

    ```bash
    export AZ_REGISTRY_NAME=registry
    export AZ_RESOURCE_GROUP=my-resource-group
    export AZ_RESOURCE_GROUP_LOCATION=westeurope
    export AZ_SUBSCRIPTION_ID=123456-123456-123456-1234567
    export AZ_SERVICE_PRINCIPAL_NAME=acr-service-principal
    export ROLE=acrpush
    export SERVER_ADDRESS=azurecr.io
    ```

2. When you communicate with Microsoft Azure for the first time, log into your Azure account. Run this command:

    ```bash
    az login
    ```

3. Create a resource group. Run:

    ```bash
    az group create --name ${AZ_RESOURCE_GROUP} --location ${AZ_RESOURCE_GROUP_LOCATION} --subscription ${AZ_SUBSCRIPTION_ID}
    ```

4. Create a Azure Container Registry. Run:

    ```bash
    az acr create --name ${AZ_REGISTRY_NAME} --resource-group ${AZ_RESOURCE_GROUP} --subscription ${AZ_SUBSCRIPTION_ID} --sku {Basic, Classic, Premium, Standard}
    ```

5. Obtain the full Azure CR ID. Run:

    ```bash
    export AZ_REGISTRY_ID=$(az acr show --name ${AZ_REGISTRY_NAME} --query id --output tsv)
    ```

6. Create the service principal with rights scoped to the Azure CR. Run:

    ```bash
    export SP_PASSWORD=$(az ad sp create-for-rbac --name http://${AZ_SERVICE_PRINCIPAL_NAME} --scopes ${AZ_REGISTRY_ID} --role ${ROLE} --query password --output tsv)
    export SP_APP_ID=$(az ad sp show --id http://${AZ_SERVICE_PRINCIPAL_NAME} --query appId --output tsv)
    ```

   Or assign the desired role to the existing service principal. Run:

    ```bash
    export SP_APP_ID=$(az ad sp show --id http://${AZ_SERVICE_PRINCIPAL_NAME} --query appId --output tsv)
    export SP_PASSWORD=$(az ad sp show --id http://${AZ_SERVICE_PRINCIPAL_NAME} --query password --output tsv)
    az role assignment create --assignee ${SP_APP_ID} --scope ${AZ_REGISTRY_ID} --role ${ROLE}
    ```

  </details>
</div>

### Override Serverless configuration

Apply the following ConfigMap with an override to a cluster or Minikube. Run:

<div tabs name="override" group="external-docker-registry">
  <details>
  <summary label="docker-hub">
  Docker Hub
  </summary>

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: serverless-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: serverless
    kyma-project.io/installation: ""
data:
  dockerRegistry.enableInternal: false
  dockerRegistry.username: "${USERNAME}"
  dockerRegistry.password: "${PASSWORD}"
  dockerRegistry.serverAddress: "${SERVER_ADDRESS}"
  dockerRegistry.registryAddress: "${REGISTRY_ADDRESS}"
EOF
```

  </details>
  <details>
  <summary label="gcr">
  GCR
  </summary>

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: serverless-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: serverless
    kyma-project.io/installation: ""
data:
  dockerRegistry.enableInternal: false
  dockerRegistry.username: "_json_key"
  dockerRegistry.password: "${GCS_KEY_JSON}"
  dockerRegistry.serverAddress: "${SERVER_ADDRESS}"
  dockerRegistry.registryAddress: "${SERVER_ADDRESS}/${PROJECT}"
EOF
```

  </details>
  <summary label="azure-cr">
  Azure CR
  </summary>

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: serverless-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: serverless
    kyma-project.io/installation: ""
data:
  dockerRegistry.enableInternal: false
  dockerRegistry.username: "${SP_APP_ID}"
  dockerRegistry.password: "${SP_PASSWORD}"
  dockerRegistry.serverAddress: "${AZ_REGISTRY_NAME}.${SERVER_ADDRESS}"
  dockerRegistry.registryAddress: "${AZ_REGISTRY_NAME}.${SERVER_ADDRESS}"
EOF
```

  </details>
</div>

> **CAUTION:** If you want to set external Docker Registry before you install Kyma, you need to manually add the ConfigMap to the `installer-config-local.yaml.tpl` (local installation) or `installer-config-production.yaml.tpl` (production installation) template located under the `installation/resources` subfolder before you run the installation script.

### Trigger installation

Trigger Kyma installation or update by labeling the Installation custom resource. Run:

```bash
kubectl -n default label installation/kyma-installation action=install
```