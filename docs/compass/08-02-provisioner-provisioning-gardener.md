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
  - Gardener service account configuration (`kubeconfig.yaml`) downloaded
  - Compass with configured Runtime Provisioner and the following [overrides](#configuration-runtime-provisioner-chart) set up:
      * Kubeconfig (`provisioner.gardener.kubeconfig`)
      * Gardener project name (`provisioner.gardener.project`)
  
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
  - Gardener service account configuration (`kubeconfig.yaml`) downloaded
  - Compass with configured Runtime Provisioner and the following [overrides](#configuration-runtime-provisioner-chart) set up:
    * Kubeconfig (`provisioner.gardener.kubeconfig`)
    * Gardener project name (`provisioner.gardener.project`)

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
  - Gardener service account configuration (`kubeconfig.yaml`) downloaded
  - Compass with configured Runtime Provisioner and the following [overrides](#configuration-runtime-provisioner-chart) set up:
    * Kubeconfig (`provisioner.gardener.kubeconfig`)
    * Gardener project name (`provisioner.gardener.project`)
  
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

  4. Make a call to the Runtime Provisioner with a **tenant** header to create a cluster on GCP.

      > **NOTE:** The Runtime Agent component (`compass-runtime-agent`) in the Kyma configuration is mandatory and the order of the components matters.                                                     
                                                                          
      ```graphql
      mutation {
        provisionRuntime(
          config: {
            runtimeInput: {
              name: "{RUNTIME_NAME}"
              description: "{RUNTIME_DESCRIPTION}"
              labels: {RUNTIME_LABELS}
            }
            clusterConfig: {
              gardenerConfig: {
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
            kymaConfig: {
              version: "1.8.0"
              components: [
                { component: "compass-runtime-agent", namespace: "compass-system" }
                {
                  component: "{KYMA_COMPONENT_NAME}"
                  namespace: "{NAMESPACE_TO_INSTALL_COMPONENT_TO}"
                  configuration: [
                    { key: "{CONFIG_PROPERTY_KEY}"
                      value: "{CONFIG_PROPERTY_VALUE}"
                      secret: {true|false} # Specifies if the property is confidential
                    }
                  ]
                }
              ]
              configuration: [
                { 
                  key: "{CONFIG_PROPERTY_KEY}"
                  value: "{CONFIG_PROPERTY_VALUE}"
                  secret: {true|false} # Specifies if the property is confidential
                }
              ]
            }
          }
        ) {
          runtimeID
          id
        }
      }
      ```
    
      A successful call returns the operation status:
    
      ```graphql
        {
          "data": {
            "provisionRuntime": {
              "runtimeID": "{RUNTIME_ID}",
              "id": "{OPERATION_ID}"
            }
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

  4. Make a call to the Runtime Provisioner with a **tenant** header to create a cluster on Azure.
        
      > **NOTE:** The Runtime Agent component (`compass-runtime-agent`) in the Kyma configuration is mandatory and the order of the components matters.                                                    
                                                                          
      ```graphql
      mutation {
        provisionRuntime(
          config: {
            runtimeInput: {
              name: "{RUNTIME_NAME}"
              description: "{RUNTIME_DESCRIPTION}"
              labels: {RUNTIME_LABELS}
            }
            clusterConfig: {
              gardenerConfig: {
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
                providerSpecificConfig: { azureConfig: { vnetCidr: "10.250.0.0/19" } }
              }
            }
            kymaConfig: {
              version: "1.8.0"
              components: [
                { component: "compass-runtime-agent", namespace: "compass-system" }
                {
                  component: "{KYMA_COMPONENT_NAME}"
                  namespace: "{NAMESPACE_TO_INSTALL_COMPONENT_TO}"
                  configuration: [
                    { key: "{CONFIG_PROPERTY_KEY}"
                      value: "{CONFIG_PROPERTY_VALUE}"
                      secret: {true|false} # Specifies if the property is confidential
                    }
                  ]
                }
              ]
              configuration: [
                { 
                  key: "{CONFIG_PROPERTY_KEY}"
                  value: "{CONFIG_PROPERTY_VALUE}"
                  secret: {true|false} # Specifies if the property is confidential
                }
              ]
            }
          }
        ) {
          runtimeID
          id
        }
      }
      ```
    
      A successful call returns the operation status:
    
      ```graphql
      {
        "data": {
          "provisionRuntime": {
            "runtimeID": "{RUNTIME_ID}",
            "id": "{OPERATION_ID}"
          }
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

  4. Make a call to the Runtime Provisioner with a **tenant** header to create a cluster on AWS.
    
      > **NOTE:** The Runtime Agent component (`compass-runtime-agent`) in the Kyma configuration is mandatory and the order of the components matters.
                                                                      
      ```graphql
      mutation {
        provisionRuntime(
          config: {
            runtimeInput: {
              name: "{RUNTIME_NAME}"
              description: "{RUNTIME_DESCRIPTION}"
              labels: {RUNTIME_LABELS}
            }
            clusterConfig: {
              gardenerConfig: {
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
                    publicCidr: "10.250.96.0/22"
                    vpcCidr: "10.250.0.0/16"
                    internalCidr: "10.250.112.0/22"
                    zone: "eu-west-1b"
                  } 
                }
              }
            }
            kymaConfig: {
              version: "1.8.0"
              components: [
                { component: "compass-runtime-agent", namespace: "compass-system" }
                {
                  component: "{KYMA_COMPONENT_NAME}"
                  namespace: "{NAMESPACE_TO_INSTALL_COMPONENT_TO}"
                  configuration: [
                    { key: "{CONFIG_PROPERTY_KEY}"
                      value: "{CONFIG_PROPERTY_VALUE}"
                      secret: {true|false} # Specifies if the property is confidential
                    }
                  ]
                }
              ]
              configuration: [
                { 
                  key: "{CONFIG_PROPERTY_KEY}"
                  value: "{CONFIG_PROPERTY_VALUE}"
                  secret: {true|false} # Specifies if the property is confidential
                }
              ]
            }
          }
        ) {
          runtimeID
          id
        }
      }
      ```
    
      A successful call returns the operation status:
    
      ```graphql
      {
        "data": {
          "provisionRuntime": {
            "runtimeID": "{RUNTIME_ID}",
            "id": "{OPERATION_ID}"
          }
        }
      }
      ```
     
  </details>
    
</div>

The operation of provisioning is asynchronous. The operation of provisioning returns the Runtime Operation Status containing the Runtime ID (`provisionRuntime.runtimeID`) and the operation ID (`provisionRuntime.id`). Use the Runtime ID to [check the Runtime Status](#tutorials-check-runtime-status). Use the provisioning operation ID to [check the Runtime Operation Status](#tutorials-check-runtime-operation-status) and verify that the provisioning was successful.

> **NOTE:** To see how to provide the labels, see [this](https://github.com/kyma-incubator/compass/blob/master/docs/compass/03-02-labeling.md) document. To see an example of label usage, go [here](https://github.com/kyma-incubator/compass/blob/master/components/director/examples/register-application/register-application.graphql). 