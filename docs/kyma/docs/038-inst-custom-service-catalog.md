---
title: Installation with custom Service Catalog deployment
type: Installation
---

You can use Kyma with a custom deployment of Service Catalog. To enable such implementation, remove Service Catalog from the list of components that install with Kyma.

## Prerequisites

- Service Catalog in version at least v0.1.28 or higher. To check the currently supported by Kyma, version of the Service Catalog, see this [file](../../../resources/service-catalog/charts/catalog/values.yaml).
  >**NOTE:** Follow [this](https://kubernetes.io/docs/tasks/service-catalog/) quick start guide to learn how to install and configure Service Catalog on a Kubernetes cluster.

- Kyma downloaded from the latest [release](https://github.com/kyma-project/kyma/releases).

## Local installation

1. Remove these lines from the [installer-cr.yaml.tpl](../../../installation/resources/installer-cr.yaml.tpl) file:
  ```
  name: "service-catalog"
  namespace: "kyma-system"
  ```
2. Follow the installation steps described in the **Install Kyma locally from the release** document.

## Cluster installation

1. Remove these lines from the [installer-cr-cluster.yaml.tpl](../../../installation/resources/installer-cr-cluster.yaml.tpl) file:
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
2. Sign in to the Kyma Console using the `admin@kyma.cx` as described in the **Install Kyma locally from the release** document.
