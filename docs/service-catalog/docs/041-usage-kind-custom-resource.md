---
title: UsageKind
type: Custom Resource
---

The `usagekinds.servicecatalog.kyma.cx` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to define which resources can be bound with the ServiceBinding and how to bind them. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd usagekinds.servicecatalog.kyma.cx -o yaml
```

## Sample custom resource

This is a sample resource that allows you to bind a given resource with the ServiceBinding. This example has a **resource** section specified as `function`. You can adjust this section to point to any other kind of resource.

```
apiVersion: servicecatalog.kyma.cx/v1alpha1
kind: UsageKind
metadata:
   name: function
spec:
   displayName: Function
   resource:
     group: kubeless.io
     kind: function
     version: v1beta1
   labelsPath: spec.deployment.spec.template.metadata.labels
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.displayName** |    **YES**   | Specifies a human-readable name of the UsageKind. |
| **spec.resource** |    **YES**   | Specifies a resource which is bound with the ServiceBinding. The target resource is specified by its resource group, kind, and version. |
| **spec.resource.group** |    **YES**   | Specifies the group of the resource. |
| **spec.resource.kind** |    **YES**   | Specifies the kind of the resource. |
| **spec.resource.version** |    **YES**   | Specifies the version of the resource. |
| **spec.labelsPath** |    **YES**   | Specifies a path to the key that contains labels which are later injected into Pods. |
