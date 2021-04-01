---
title: Bind an Application to a Namespace
type: Tutorials
---

This guide shows you how to bind an Application to a Namespace in Kyma. To execute the binding, create an ApplicationMapping custom resource in the cluster. Follow the instructions to bind your Application to a desired Namespace.

## Prerequisites

- Application created in your cluster
- Namespace to which you want to bind the Application

## Steps

1. Export the Namespace to which you want to bind the Application.
      
      ```bash
      export NAMESPACE={YOUR_NAMESPACE}

2. List all Applications bound to the Namespace.

  ```bash
  kubectl get am -n $NAMESPACE
  ```

3. Bind an Application to the Namespace. Create an ApplicationMapping custom resource and apply it to the cluster.

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: ApplicationMapping
  metadata:
    name: {NAME_OF_APP_TO_BIND}
    namespace: $NAMESPACE
  EOF
  ```

4. Check if the operation succeeded. List all Namespaces to which your Application is bound.

  ```bash
  kubectl get am --all-namespaces -o jsonpath='{range .items[?(@.metadata.name=="{NAME_OF_YOUR_APP}")]}{@.metadata.namespace}{""}{end}'
  ```
