---
title: Runtime Provisioner
type: Details
---

The Runtime Provisioner is a Compass component responsible for provisioning, installing, and deprovisioning clusters with Kyma (Kyma Runtimes). The relationship between clusters and Runtimes is 1:1.

It allows you to provision the clusters in the following ways:
- [through Gardener](#tutorials-provision-clusters-through-gardener) on:
    * GCP
    * Microsoft Azure
    * Amazon Web Services (AWS).

During the operation of provisioning, you can pass a list of Kyma components you want installed on the provisioned Runtime with their custom configuration, as well as a custom Runtime configuration. 

Note that the operations of provisioning and deprovisioning are asynchronous. The operation of provisioning returns the Runtime Operation Status containing the Runtime ID and the operation ID. The operation of deprovisioning returns the operation ID. You can use the operation ID to [check the Runtime Operation Status](#tutorials-check-runtime-operation-status) and the Runtime ID to [check the Runtime Status](#tutorials-check-runtime-status).
  
The Runtime Provisioner exposes an API to manage cluster provisioning, installation, and deprovisioning. 

Find the specification of the API [here](https://github.com/kyma-incubator/compass/blob/master/components/provisioner/pkg/gqlschema/schema.graphql).
    
To access the Runtime Provisioner, forward the port that the GraphQL Server is listening on:

```bash
kubectl -n compass-system port-forward svc/compass-provisioner 3000:3000
```

When making a call to the Runtime Provisioner, make sure to attach a tenant header to the request.