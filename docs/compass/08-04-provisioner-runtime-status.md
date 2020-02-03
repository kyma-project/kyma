---
title: Check Runtime Status
type: Tutorials
---

This tutorial shows how to check the Runtime status.

## Steps

Make a call to the Runtime Provisioner with a **tenant** header to check the Runtime status. Pass the Runtime ID as `id`. 

```graphql
query { runtimeStatus(id: "{RUNTIME_ID}") {
    lastOperationStatus {
      id operation state message runtimeID 
  	} 
    runtimeConnectionStatus { 
      status errors {
        message
      } 
    } 

    runtimeConfiguration {
      clusterConfig {
        __typename ... on GCPConfig {
          bootDiskSizeGB name numberOfNodes kubernetesVersion projectName machineType zone region }
          ... on GardenerConfig { 
          name workerCidr region diskType maxSurge nodeCount volumeSizeGB projectName machineType targetSecret 
          autoScalerMin autoScalerMax provider maxUnavailable kubernetesVersion }
      }
      kymaConfig {
        version  components {
          component
          namespace configuration {
            key
            value
            secret
          }
        }
        configuration {
          key value secret
        }
      }
    	kubeconfig
    } 
  } 
}
```

An example response for a successful request looks like this:

```graphql
{
  "data": {
    "runtimeStatus": {
      "lastOperationStatus": {
        "id": "20ed1cfb-7407-4ec5-89af-c550eb0fce49",
        "operation": "Provision",
        "state": "Succeeded",
        "message": "Operation succeeded.",
        "runtimeID": "b70accda-4008-466c-96ec-9b42c2cfd264"
      },
      "runtimeConnectionStatus": {
        "status": "Pending",
        "errors": null
      },
      "runtimeConfiguration": {
        "clusterConfig": {CLUSTER_CONFIG},
        "kymaConfig": {
          "version": "1.8.0",
          "components": [{COMPONENTS_LIST}]
        },
        "kubeconfig": {KUBECONFIG}
      }
    }
  }
}
``` 