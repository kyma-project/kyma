---
title: Bind an Application to a Namespace
type: Tutorials
---

This guide shows you how to bind an Application to a Namespace in Kyma. To execute the binding, create an ApplicationMapping custom resource in the cluster. Follow the instructions to bind your Application to the `production` Namespace.

## Prerequisites

- An Application created in your cluster

## Steps

1. List all Applications bound to the `production` Namespace.
  ```
  kubectl get em -n production
  ```

2. Bind an Application to a Namespace. Create an ApplicationMapping custom resource and apply it to the cluster.

  ```
  cat <<EOF | kubectl apply -f -
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: ApplicationMapping
  metadata:
    name: {NAME_OF_APP_TO_BIND}
    namespace: production
  EOF
  ```

3. Check if the operation succeeded. List all Namespaces to which your Application is bound.
  ```
  kubectl get em --all-namespaces -o jsonpath='{range .items[?(@.metadata.name=="{NAME_OF_YOUR_APP}")]}{@.metadata.namespace}{""}{end}'
  ```
