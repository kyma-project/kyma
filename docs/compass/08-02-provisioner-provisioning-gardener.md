---
title: Provision clusters through Gardener
type: Tutorials
---

This tutorial shows how to provision clusters with Kyma Runtimes on Google Cloud Platform (GCP), Microsoft Azure, and Amazon Web Services (AWS) using [Gardener](https://dashboard.garden.canary.k8s.ondemand.com).

## Prerequisites

<div tabs name="Prerequisites" group="Provisioning-Gardener">
  <details>
  <summary label="GCP">
  GCP
  </summary>
  
  - Existing project on GCP
  - Existing project on Gardener
  - Service account for GCP with the following roles:
      * Service Account Admin
      * Service Account Token Creator
      * Service Account User
      * Compute Admin
  - Key generated for your service account, downloaded in the JSON format   
  </details>
  
  <details>
  <summary label="Azure">
  Azure
  </summary>
  
  - Existing project on Gardener
  - Valid Azure subscription with the Contributor role and the subscription ID 
  - Existing App registration on Azure with the following credentials:
    * Application ID (Client ID)
    * Directory ID (Tenant ID)
    * Client secret (application password)

  </details>
  
  <details>
  <summary label="AWS">
  AWS
  </summary>
  
  - Existing project on Gardener
  - AWS account with added AWS IAM policy for Gardener
  - Access key created for your AWS user with the following credentials:
    * Secrete Access Key
    * Access Key ID
  
  > **NOTE:** To get the AWS IAM policy, access your project on Gardener, navigate to the **Secrets** tab, click on the help icon on the AWS card, and copy the JSON policy. 
    
  </details>
</div>

> **NOTE:** To access the Runtime Provisioner, forward the port on which the GraphQL Server is listening.
   
## Steps

<div tabs name="Provisioning" group="Provisioning-Gardener">
  <details>
  <summary label="GCP">
  GCP
  </summary>

  To provision Kyma Runtime on GCP, follow these steps:

  1. Access your project on [Gardener](https://dashboard.garden.canary.k8s.ondemand.com).

  2. In the **Secrets** tab, add a new Google Secret for GCP. Use the `json` file with the service account key you downloaded from GCP.

  3. In the **Members** tab, create a service account for Gardener. 

  4. Download the service account configuration (`kubeconfig.yaml`) and use it to create a Secret in the `compass-system` Namespace with the key `credentials` and the value encoded with base64.

  5. Make a call to the Runtime Provisioner to create a cluster on GCP.

      > **NOTE:** The cluster name must start with a lowercase letter followed by up to 19 lowercase letters, numbers, or hyphens, and cannot end with a hyphen.                                                                 
                                                                          
      ```graphql
      mutation { 
        provisionRuntime(
          id:"61d1841b-ccb5-44ed-a9ec-45f70cd2b0d6" config: {
            clusterConfig: {
              gardenerConfig: {
                name: "{CLUSTER_NAME}" 
                projectName: "{GARDENER_PROJECT_NAME}" 
                kubernetesVersion: "1.15.4"
                diskType: "pd-standard"
                volumeSizeGB: 30
                nodeCount: 3
                machineType: "n1-standard-4"
                region: "europe-west4"
                provider: "gcp"
                seed: "gcp-eu1"
                targetSecret: "{GARDENER_GCP_SECRET_NAME}"
                workerCidr: "10.250.0.0/19"
                autoScalerMin: 2
                autoScalerMax: 4
                maxSurge: 4
                maxUnavailable: 1
                providerSpecificConfig: { gcpConfig: { zone: "europe-west4-a" } }
              }
            }
            kymaConfig: { version: "1.5", modules: Backup }
            credentials: {
              secretName: "{GARDENER_SERVICE_ACCOUNT_CONFIGURATION_SECERT_NAME}" 
            }
          }
        )
      }
      ```
    
      A successful call returns the ID of the provisioning operation:
    
      ```graphql
      {
        "data": {
          "provisionRuntime": "7a8dc760-812c-4a35-a5fe-656a648ee2c8"
        }
      }
      ```
    
  </details>

  <details>
  <summary label="Azure">
  Azure
  </summary>

  To provision Kyma Runtime on Azure, follow these steps:

  1. Access your project on [Gardener](https://dashboard.garden.canary.k8s.ondemand.com).

  2. In the **Secrets** tab, add a new Azure Secret. Use the credentials you got from Azure.

  3. In the **Members** tab, create a service account for Gardener. 

  4. Download the service account configuration (`kubeconfig.yaml`) and use it to create a Secret in the `compass-system` Namespace with the key `credentials` and the value encoded with base64.

  5. Make a call to the Runtime Provisioner to create a cluster on Azure.

      > **NOTE:** To access the Runtime Provisioner, forward the port on which the GraphQL Server is listening.
    
      > **NOTE:** The cluster name must start with a lowercase letter followed by up to 19 lowercase letters, numbers, or hyphens, and cannot end with a hyphen.                                                                  
                                                                          
      ```graphql
      mutation { 
        provisionRuntime(
          id:"61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3" config: {
            clusterConfig: {
              gardenerConfig: {
                name: "{CLUSTER_NAME}"
                projectName: "{GARDENER_PROJECT_NAME}"
                kubernetesVersion: "1.15.4"
                diskType: "Standard_LRS"
                volumeSizeGB: 35
                nodeCount: 3
                machineType: "Standard_D2_v3"
                region: "westeurope"
                provider: "azure"
                seed: "az-eu1"
                targetSecret: "{GARDENER_AZURE_SECRET_NAME}"
                workerCidr: "10.250.0.0/19"
                autoScalerMin: 2
                autoScalerMax: 4
                maxSurge: 4
                maxUnavailable: 1
                providerSpecificConfig: {  azureConfig: { vnetCidr: "10.250.0.0/19" } }
              }
            }
            kymaConfig: { version: "1.5", modules: Backup }
            credentials: { secretName: "{GARDENER_SERVICE_ACCOUNT_CONFIGURATION_SECRET_NAME}" }
          }
        )
      }
      ```
    
      A successful call returns the ID of the provisioning operation:
    
      ```graphql
      {
        "data": {
          "provisionRuntime": "af0c8122-27ee-4a36-afa5-6e26c39929f2"
        }
      }
      ```
    
  </details>
  
  <details>
  <summary label="AWS">
  AWS
  </summary>
      
  To provision Kyma Runtime on AWS, follow these steps:
    
  1. Access your project on [Gardener](https://dashboard.garden.canary.k8s.ondemand.com).
  
  2. In the **Secrets** tab, add a new AWS Secret. Use the credentials you got from AWS.
    
  3. In the **Members** tab, create a service account for Gardener. 

  4. Download the service account configuration (`kubeconfig.yaml`) and use it to create a Secret in the `compass-system` Namespace with the key `credentials` and the value encoded with base64.

  5. Make a call to the Runtime Provisioner to create a cluster on AWS.

      > **NOTE:** To access the Runtime Provisioner, forward the port on which the GraphQL Server is listening.
    
      > **NOTE:** The cluster name must start with a lowercase letter followed by up to 19 lowercase letters, numbers, or hyphens, and cannot end with a hyphen.                                                                  
                                                                          
      ```graphql
      mutation { 
        provisionRuntime(
          id:"61d1841b-ccb5-44ed-a9ec-15f70cd2b0d2" 
          config: {
            clusterConfig: {
              gardenerConfig: {
                name: "{CLUSTER_NAME}"
                projectName: "{GARDENER_PROJECT_NAME}"
                kubernetesVersion: "1.15.4"
                diskType: "gp2"
                volumeSizeGB: 35
                nodeCount: 3
                machineType: "m4.2xlarge"
                region: "eu-west-1"
                provider: "aws"
                seed: "aws-eu1"
                targetSecret: "{GARDENER_AWS_SECRET_NAME}"
                workerCidr: "10.250.0.0/19"
                autoScalerMin: 2
                autoScalerMax: 4
                maxSurge: 4
                maxUnavailable: 1
                providerSpecificConfig: { 
                  awsConfig: {
                    publicCidr: "10.250.96.0/22",
                    vpcCidr:         "10.250.0.0/16",
                    internalCidr:   "10.250.112.0/22",
                    zone:            "eu-west-1b",
                  }
                }
              }
            }
            kymaConfig: { version: "1.5", modules: Backup }
            credentials: { secretName: "{GARDENER_SERVICE_ACCOUNT_CONFIGURATION_SECRET_NAME}" }
          }
        )
      }
      ```
    
      A successful call returns the ID of the provisioning operation:
    
      ```graphql
      {
        "data": {
          "provisionRuntime": "55dab98f-4efc-4afa-81df-b40ae2de146a"
        }
      }
      ```
  </details>
    
</div>

The operation of provisioning is asynchronous. Use the provisioning operation ID (`provisionRuntime`) to [check the Runtime Operation Status](#tutorials-check-runtime-operation-status) and verify that the provisioning was successful. Use the Runtime ID (`id`) to [check the Runtime Status](#tutorials-check-runtime-status). 