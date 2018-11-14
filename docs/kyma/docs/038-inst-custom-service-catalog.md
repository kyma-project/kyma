---
title: Installation with custom Service Catalog deployment
type: Installation
---

You can use Kyma with a custom Service Catalog deployment. To enable such implementation, remove the Service Catalog from the list of components that install with Kyma.

## Prerequisites

- The Service Catalog in the Kyma-supported version . To check the currently supported version of the Service Catalog, see the value of the **image** parameter in [this](https://github.com/kyma-project/kyma/tree/master/resources/service-catalog/charts/catalog/values.yaml) file.
>**NOTE:** Follow [this](https://kubernetes.io/docs/tasks/service-catalog/) guide to learn how to install and configure Service Catalog on a Kubernetes cluster.

- Kyma latest [release](https://github.com/kyma-project/kyma/releases).

## Local installation

1. Remove these lines from the [installer-cr.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr.yaml.tpl) file:
  ```
  name: "service-catalog"
  namespace: "kyma-system"
  ```
2. Follow the installation steps described in the **Install Kyma locally from the release** document.

## Cluster installation

1. Remove these lines from the [installer-cr-cluster.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster.yaml.tpl) file:
  ```
  name: "service-catalog"
  namespace: "kyma-system"
  ```
2. Follow the installation steps described in the **Install Kyma on a GKE cluster** document.

## Verify the installation

1. Check if all Pods are running in the `kyma-system` Namespace:
  ```
  kubectl get pods -n kyma-system
  ```
2. Sign in to the Kyma Console using the `admin@kyma.cx` login as described in the **Install Kyma locally from the release** document.
