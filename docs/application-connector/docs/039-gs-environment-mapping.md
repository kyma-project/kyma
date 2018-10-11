---
title: Bind a Remote Environment to an Environment
type: Getting Started
---

This guide shows you how to bind a Remote Environment (RE) to an Environment in Kyma. To execute the binding, create an EnvironmentMapping Custom Resource in the cluster. Follow the instructions to bind your Remote Environment to the `production` Environment.

## Prerequisites

To complete this guide, your cluster must have at least one Remote Environment created.

## Steps

1. List all Remote Environments bound to the `production` Environment:
  ```
  kubectl get em -n production
  ```

2. Create an EnvironmentMapping Custom Resource (CR) for your environment and save it to a `mapping-prod.yaml` file. Follow this template:
  ```
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: EnvironmentMapping
  metadata:
    name: {NAME_OF_RE_TO_BIND}
    namespace: production
  ```

3. Create the CR in the cluster:  
  ```
  kubectl apply -f mapping-prod.yaml
  ```

4. Check if the operation is successful. List all Environments to which your RE is bound:
  ```
  kubectl get em --all-namespaces -o jsonpath='{range .items[?(@.metadata.name=="{NAME_OF_YOUR_RE}")]}{@.metadata.namespace}{"\n"}{end}'
  ```
