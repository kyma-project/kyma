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

2. Bind a RE to an Environment. Run this command to create an EnvironmentMapping Custom Resource and apply it to the cluster:

  ```
  cat <<EOF | kubectl apply -f -
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: EnvironmentMapping
  metadata:
    name: {NAME_OF_RE_TO_BIND}
    namespace: production
  EOF
  ```

3. Check if the operation is successful. List all Environments to which your RE is bound:
  ```
  kubectl get em --all-namespaces -o jsonpath='{range .items[?(@.metadata.name=="{NAME_OF_YOUR_RE}")]}{@.metadata.namespace}{""}{end}'
  ```
