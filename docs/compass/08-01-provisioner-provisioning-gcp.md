---
title: Provision clusters on Google Cloud Platform
type: Tutorials
---

This tutorial shows how to provision clusters with Kyma Runtimes on Google Cloud Platform (GCP).

## Prerequisites

- Existing project on GCP
- Service account on GCP with the following roles:
    * Compute Admin
    * Kubernetes Engine Admin
    * Kubernetes Engine Cluster Admin
    * Service Account User
- Key generated for your service account, downloaded in the JSON format
- Secret from the service account key created in the `compass-system` Namespace, with the key `credentials` and the value encoded with base64

## Steps

To provision Kyma Runtime, make a call to the Runtime Provisioner with this example mutation:

> **NOTE:** To access the Runtime Provisioner, forward the port on which the GraphQL Server is listening.

> **NOTE:** The cluster name must start with a lowercase letter followed by up to 39 lowercase letters, numbers, or hyphens, and cannot end with a hyphen.

```graphql
mutation { 
  provisionRuntime(
    id:"309051b6-0bac-44c8-8bae-3fc59c12bb5c" 
    config: {
      clusterConfig: {
        gcpConfig: {
          name: "{CLUSTER_NAME}"
          projectName: "{GCP_PROJECT_NAME}"
          kubernetesVersion: "1.13"
          bootDiskSizeGB: 30
          numberOfNodes: 1
          machineType: "n1-standard-4"
          region: "europe-west3-a"
         }
      }
      kymaConfig: {
        version: "1.5"
        modules: Backup
      }
      credentials: {
        secretName: "{SECRET_NAME}"
      }
    }
  )
}
```

A successful call returns the ID of the provisioning operation:

```graphql
{
  "data": {
    "provisionRuntime": "e9c9ed2d-2a3c-4802-a9b9-16d599dafd25"
  }
}
```

The operation of provisioning is asynchronous. Use the provisioning operation ID (`provisionRuntime`) to [check the Runtime Operation Status](08-03-provisioner-runtime-operation-status.md) and verify that the provisioning was successful. Use the Runtime ID (`id`) to [check the Runtime Status](08-04-provisioner-runtime-status.md). 