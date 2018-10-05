---
title: Activate a RemoteEnvironment using EnvironmentMapping
type: Getting Started
---

This Getting Started guide shows you how to bind Remote Environments to Environment in the command line. For the Custom Resource Definition, see the `environment-mapping.crd.yaml` file under the `/resources/cluster-essentials/templates/` directory.
An instance of the EnvironmentMapping enables the RemoteEnvironment with the same name in a given Namespace. In this example, the EnvironmentMapping enables the `ec-prod` remote environment in the `production` Namespace:

```yaml
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: EnvironmentMapping
metadata:
  name: ec-prod
  namespace: production
```

## Prerequisites

You need to have RE created.

## Details

Follow these steps to complete the Getting Started guide:

1. List all RemoteEnvironments enabled in the `production` Environment:
    ```bash
    > kubectl get em -n production
    
    No resources found.
    ```

2. Enable this RemoteEnvironment in the `production` Environment:

    Create a file mapping-prod.yaml:
    
    ```yaml
    apiVersion: applicationconnector.kyma-project.io/v1alpha1
    kind: EnvironmentMapping
    metadata:
      name: ec-prod
      namespace: production
    ```
    
    and apply it:
    
    ```bash
    > kubectl apply -f mapping-prod.yaml
    
    environmentmapping "ec-prod" created
    ```
      
3. List all RemoteEnvironments enabled in the `production` Environment again:
    
    ```bash
    > kubectl get em -n production
    NAME      AGE
    ec-prod   40s
    ```
    
4. Unbind RE from `production` Environment:
    
    ```bash
    > kubectl delete em ec-prod -n production
    ```
    
5. List all environments where `ec-prod` is enabled
    
    ```bash
    > kubectl get em --all-namespaces -o jsonpath='{range .items[?(@.metadata.name=="ec-prod")]}{@.metadata.namespace}{"\n"}{end}'
    production
    ```

