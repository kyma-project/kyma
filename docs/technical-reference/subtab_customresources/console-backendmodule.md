---
title: BackendModule
type: Custom Resource
---

The `backendmodules.ui.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to enable Console Backend Service modules.

To get the up-to-date CRD and show the output in the `yaml` format, run this command:

``` bash
kubectl get crd backendmodules.ui.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample CR that enables the `servicecatalog` module in the Console Backend Service:

``` yaml
apiVersion: ui.kyma-project.io/v1alpha1
kind: BackendModule
metadata:
  name: servicecatalog
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR. It must be the same as the name of a given Console Backend Service module. |

## Related resources and components

These components use this CR:

| Component   |   Description |
|----------|------|
| Console Backend Service |  The component reacts to every action of adding or deleting the BackendModule custom resource and enables or disables a given Console Backend Service module accordingly. |
