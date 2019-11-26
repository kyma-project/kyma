---
title: Check Runtime Status
type: Tutorials
---

This tutorial shows how to check the Runtime status.

## Steps

Make a call to the Runtime Provisioner to check the Runtime status. Pass the Runtime ID as `id`. 

```graphql
query { runtimeStatus(id: "309051b6-0bac-44c8-8bae-3fc59c12bb5c") {
  runtimeConnectionStatus {
    status errors {
      message
    } 
  } 
  lastOperationStatus {
    message operation state runtimeID id
  } 
  runtimeConfiguration {
    kubeconfig kymaConfig {
      version modules 
    } clusterConfig {
      __typename ... on GCPConfig { bootDiskSizeGB name numberOfNodes kubernetesVersion projectName machineType zone region }
      ... on GardenerConfig { name workerCidr region diskType maxSurge nodeCount volumeSizeGB projectName machineType targetSecret autoScalerMin autoScalerMax provider maxUnavailable kubernetesVersion providerSpecificConfig
      { __typename 
        ... on GCPProviderConfig { zone } 
        ... on AzureProviderConfig {vnetCidr}
                ... on AWSProviderConfig {zone internalCidr vpcCidr publicCidr}      
      }
      } } } } }
```

An example response for a successful request looks like this:

```graphql
{
  "data": {
    "runtimeStatus": {
      "runtimeConnectionStatus": {
        "status": "Pending",
        "errors": null
      },
      "lastOperationStatus": {
        "message": "Operation succeeded.",
        "operation": "Provision",
        "state": "Succeeded",
        "runtimeID": "309051b6-0bac-44c8-8bae-3fc59c12bb5c",
        "id": "e9c9ed2d-2a3c-4802-a9b9-16d599dafd25"
      },
      "runtimeConfiguration": {
        "kubeconfig": "{KUBECONFIG}",
        "kymaConfig": {
          "version": "1.5",
          "modules": [
            "Backup"
          ]
        },
        "clusterConfig": {CLUSTER_CONFIG}
      }
    }
  }
}
``` 