---
title: Overview
type: Overview
---

>**NOTE:** To provision this class, first you must create a Secret. Read the following document to learn how.

The Microsoft Azure Service Broker is an open source, Open Service Broker-compatible API server that provisions managed services in the Microsoft Azure public cloud.

## Create a Secret

### Prerequisites
 
Install [the Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest). 

### Steps
Follow these steps to create a proper Kubernetes Secret with all necessary data.
1. Run `az login` and follow the instructions in the command output to authorize `az` to use your account.
2. List your Azure subscriptions and choose the correct one:
```bash
az account list -o table
```
Then, save the subscription ID in an environment variable:
```bash
export AZURE_SUBSCRIPTION_ID="{SubscriptionId}"
```
3. Create a resource group to contain the resources you will be creating.
```bash
az group create --name {group name} --location eastus
```
You can also use one of the existing resource groups:
```bash
az group list -o table
```
4. Create a service principal with RBAC enabled:
```bash
az ad sp create-for-rbac --name {group name} -o table
```
Save the following values:
```yaml
export AZURE_TENANT_ID={Tenant}
export AZURE_CLIENT_ID={AppId}
export AZURE_CLIENT_SECRET={Password}
```
5. Create a Secret:
```bash
kubectl create secret generic azure-broker-data -n {namespace} \
        --from-literal=subscription_id=$AZURE_SUBSCRIPTION_ID \
        --from-literal=tenant_id=$AZURE_TENANT_ID \
        --from-literal=client_secret=$AZURE_CLIENT_SECRET \
        --from-literal=client_id=$AZURE_CLIENT_ID
```

>**NOTE:** You can provision only one instance of the Azure Service Broker in each Namespace.