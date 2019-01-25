---
title: BackendModule
type: Custom Resource
---

The `backendmodule.ui.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to enable UI API Layer modules.

To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd backendmodules.ui.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample CR that enables the `servicecatalog` module in the UI API Layer:

```
apiVersion: ui.kyma-project.io/v1alpha1
kind: BackendModule
metadata:
  name: servicecatalog
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. It must be the same as the name of a given UI API Layer module. |

## Related resources and components

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| UI API Layer |  The component reacts to every action of adding or deleting the BackendModule custom resource and enables or disables a given UI API Layer module accordingly. |
