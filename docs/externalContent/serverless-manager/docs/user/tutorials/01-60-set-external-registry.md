# Set an External Docker Registry

By default, you install Kyma with Serverless that uses the internal Docker registry running in a cluster. This tutorial shows how to override this default setup with an external Docker registry from one of these cloud providers:

- [Docker Hub](https://hub.docker.com/)
- [Google Artifact Registry (GAR)](https://cloud.google.com/artifact-registry)
- [Azure Container Registry (ACR)](https://azure.microsoft.com/en-us/services/container-registry/)

> [!WARNING]
> Function images are not cached in the Docker Hub. The reason is that this registry is not compatible with the caching logic defined in [Kaniko](https://cloud.google.com/cloud-build/docs/kaniko-cache) that Serverless uses for building images.

## Prerequisites

<Tabs>
<Tab name="Docker Hub">

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
</Tab>
<Tab name="GAR">

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [Google Cloud Platform (GCP)](https://cloud.google.com) project
</Tab>
<Tab name="ACR">

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure)
- [Microsoft Azure](http://azure.com) subscription
</Tab>
</Tabs>

## Steps

### Create Required Cloud Resources

<Tabs>
<Tab name="Docker Hub">

1. Run the `export {VARIABLE}={value}` command to set up these environment variables, where:

    - **USER_NAME** is the name of the account in the Docker Hub.
    - **PASSWORD** is the password for the account in the Docker Hub.
    - **SERVER_ADDRESS** is the server address of the Docker Hub. At the moment, Kyma only supports the `https://index.docker.io/v1/` server address.
    - **REGISTRY_ADDRESS** is the registry address in the Docker Hub.

    > [!TIP]
    > Usually, the Docker registry address is the same as the account name.

    Example:

    ```bash
    export USER_NAME=kyma-rocks
    export PASSWORD=admin123
    export SERVER_ADDRESS=https://index.docker.io/v1/
    export REGISTRY_ADDRESS=kyma-rocks
    ```
</Tab>
<Tab name="GAR">

To use GAR, create a Google service account that has a private key and the **Storage Admin** role permissions. Follow these steps:

1. Run the `export {VARIABLE}={value}` command to set up these environment variables, where:

    - **SA_NAME** is the name of the service account.
    - **SA_DISPLAY_NAME** is the display name of the service account.
    - **PROJECT** is the GCP project ID.
    - **SECRET_FILE** is the path to the private key.
    - **ROLE** is the **Storage Admin** role bound to the service account.
    - **SERVER_ADDRESS** is the server address of the Docker registry.

    Example:

    ```bash
    export SA_NAME=my-service-account
    export SA_DISPLAY_NAME=service-account
    export PROJECT=test-project-012345
    export SECRET_FILE=my-private-key-path
    export ROLE=roles/storage.admin
    export SERVER_ADDRESS=gar.io
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
    export GCS_KEY_JSON=$(< "$SECRET_FILE" base64 | tr -d '\n')
    ```
</Tab>
<Tab name="ACR">

Create an ACR and a service principal. Follow these steps:

1. Run the `export {VARIABLE}={value}` command to set up these environment variables, where:

    - **AZ_REGISTRY_NAME** is the name of the ACR.
    - **AZ_RESOURCE_GROUP** is the name of the resource group.
    - **AZ_RESOURCE_GROUP_LOCATION** is the location of the resource group.
    - **AZ_SUBSCRIPTION_ID** is the ID of the Azure subscription.
    - **AZ_SERVICE_PRINCIPAL_NAME** is the name of the Azure service principal.
    - **ROLE** is the **acrpush** role bound to the service principal.
    - **SERVER_ADDRESS** is the server address of the Docker registry.

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

4. Create an ACR. Run:

    ```bash
    az acr create --name ${AZ_REGISTRY_NAME} --resource-group ${AZ_RESOURCE_GROUP} --subscription ${AZ_SUBSCRIPTION_ID} --sku {Basic, Classic, Premium, Standard}
    ```

5. Obtain the full ACR ID. Run:

    ```bash
    export AZ_REGISTRY_ID=$(az acr show --name ${AZ_REGISTRY_NAME} --query id --output tsv)
    ```

6. Create a service principal with rights scoped to the ACR. Run:

    ```bash
    export SP_PASSWORD=$(az ad sp create-for-rbac --name http://${AZ_SERVICE_PRINCIPAL_NAME} --scopes ${AZ_REGISTRY_ID} --role ${ROLE} --query password --output tsv)
    export SP_APP_ID=$(az ad sp show --id http://${AZ_SERVICE_PRINCIPAL_NAME} --query appId --output tsv)
    ```

   Alternatively, assign the desired role to the existing service principal. Run:

    ```bash
    export SP_APP_ID=$(az ad sp show --id http://${AZ_SERVICE_PRINCIPAL_NAME} --query appId --output tsv)
    export SP_PASSWORD=$(az ad sp show --id http://${AZ_SERVICE_PRINCIPAL_NAME} --query password --output tsv)
    az role assignment create --assignee ${SP_APP_ID} --scope ${AZ_REGISTRY_ID} --role ${ROLE}
    ```
</Tab>
</Tabs>

### Override Serverless Configuration

Prepare yaml file with overrides that match your Docker registry provider:

<Tabs>
<Tab name="Docker Hub">

```bash
cat > docker-registry-overrides.yaml <<EOF
serverless:
  dockerRegistry:
    enableInternal: false
    username: "${USER_NAME}"
    password: "${PASSWORD}"
    serverAddress: "${SERVER_ADDRESS}"
    registryAddress: "${REGISTRY_ADDRESS}"
EOF
```
</Tab>
<Tab name="GAR">

```bash
cat > docker-registry-overrides.yaml <<EOF
serverless:
  dockerRegistry:
    enableInternal: false
    username: "_json_key"
    password: "${GCS_KEY_JSON}"
    serverAddress: "${SERVER_ADDRESS}"
    registryAddress: "${SERVER_ADDRESS}/${PROJECT}"
EOF
```
</Tab>
<Tab name="ACR">

```bash
cat > docker-registry-overrides.yaml <<EOF
serverless:
  dockerRegistry:
    enableInternal: false 
    username: "${SP_APP_ID}" 
    password: "${SP_PASSWORD}" 
    serverAddress: "${AZ_REGISTRY_NAME}.${SERVER_ADDRESS}" 
    registryAddress: "${AZ_REGISTRY_NAME}.${SERVER_ADDRESS}" 
EOF
```
</Tab>
</Tabs>

> [!WARNING]
> If you want to set an external Docker registry before you install Kyma, manually apply the Secret to the cluster before you run the installation script.

### Apply Configuration

Deploy Kyma with different configuration for Docker registry . Run:

```bash
kyma deploy --values-file docker-registry-overrides.yaml
```

> [!NOTE]
> To learn more, read about [changing Kyma configuration](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/03-change-kyma-config-values).
