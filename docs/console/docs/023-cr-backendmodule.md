---
title: BackendModule
type: Custom Resource
---

The BackendModule CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to enable UI API Layer modules. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd backendmodules.ui.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that enables `servicecatalog` module in UI API Layer.

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
| **metadata.name** |    **YES**   | Specifies the name of the CR. It must match the name of one of UI API Layer modules. |


## Related resources and components

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| UI API Layer |  The component reacts when a BackendModule custom resource is added or deleted. It enables or disables proper UI API Layer module accordingly to the change. |