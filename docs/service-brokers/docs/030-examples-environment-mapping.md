---
title: EnvironmentMapping tutorial
type: Examples
---

This tutorial shows how to perform operations on remote environments in the command line. For the custom resource definition, see the [environment-mapping.crd.yaml](../../../resources/cluster-essentials/templates/environment-mapping.crd.yaml) file.

An instance of the EnvironmentMapping enables the RemoteEnvironment with the same name in a given Namespace. In this example, the EnvironmentMapping enables the `ec-prod` remote environment in the `production` Namespace:

```yaml
apiVersion: remoteenvironment.kyma.cx/v1alpha1
kind: EnvironmentMapping
metadata:
  name: ec-prod
  namespace: production
```

## Prerequisites

To follow this tutorial, run Kyma locally. For information on how to deploy Kyma locally, see the [local installation](../../kyma/docs/031-gs-local-installation.md) getting started guide.

## Details

Follow these steps to complete the tutorial:
1. List all RemoteEnvironments enabled in the `production` environment:
    ```bash
    > kubectl get em -n production
    No resources found.
    ```
2. Create a RemoteEnvironment:
    ```bash
    > kubectl apply -f docs/assets/crd/remote-environment-prod.yaml
    remoteenvironment "ec-prod" created
    ```
3. Enable this RemoteEnvironment in the `production` environment:
    ```bash
    > kubectl apply -f docs/assets/crd/mapping-prod.yaml
    environmentmapping "ec-prod" created
    ```  
4. List all RemoteEnvironments enabled in the `production` environment again:
    ```bash
    > kubectl get em -n production
    NAME      AGE
    ec-prod   40s
    ```
5. List all environments where `ec-prod` is enabled:
    ```bash
    > kubectl get em --all-namespaces -o jsonpath='{range .items[?(@.metadata.name=="ec-prod")]}{@.metadata.namespace}{"\n"}{end}'
    production
    ```
6. Disable `ec-prod` in the `production` environment:
    ```bash
    > kubectl delete -f docs/assets/crd/mapping-prod.yaml
    environmentmapping "ec-prod" deleted
    ```
7. List all environments where `ec-prod` is enabled:
    ```bash
    > kubectl get em --all-namespaces -o jsonpath='{range .items[?(@.metadata.name=="ec-prod")]}{@.metadata.namespace}{"\n"}{end}'
    ```
8. Delete all created resources:
    ```bash
    > kubectl delete -f docs/assets/crd/remote-environment-prod.yaml
    remoteenvironment "ec-prod" deleted
    ```
