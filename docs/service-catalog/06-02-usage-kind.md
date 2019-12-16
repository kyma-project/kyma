---
title: UsageKind
type: Custom Resource
---

The `usagekinds.servicecatalog.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define which resources can be bound with the ServiceBinding and how to bind them. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd usagekinds.servicecatalog.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that allows you to bind a given resource with the ServiceBinding. This example has a **resource** section specified as `function`. You can adjust this section to point to any other kind of resource.

```yaml
apiVersion: servicecatalog.kyma-project.io/v1alpha1
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

| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** |    Yes   | Specifies the name of the CR. |
| **spec.displayName** |    Yes   | Provides a human-readable name of the UsageKind. |
| **spec.resource** |    Yes   | Specifies a resource which is bound with the ServiceBinding. The target resource is specified by its resource group, kind, and version. |
| **spec.resource.group** |    Yes   | Specifies the group of the resource. |
| **spec.resource.kind** |    Yes   | Specifies the kind of the resource. |
| **spec.resource.version** |    Yes   | Specifies the version of the resource. |
| **spec.labelsPath** |    Yes   | Specifies a path to the key that contains labels which are later injected into Pods. |

## Related resources and components

These are the resources related to this CR:

| Custom resource   |   Description |
|----------|------|
| [ServiceBindingUsage](#custom-resource-servicebindingusage) |  Contains the reference to the UsageKind. |

These components use this CR:

| Component   |   Description |
|----------|------|
| Service Binding Usage Controller |  Uses the UsageKind **spec.resource** and **spec.labelsPath** parameters to find a resource and a path to which it should inject Secrets. |
| Console Backend Service |  Exposes the given CR to the Console UI. |

## RBAC settings

The administrator who adds the UsageKind must take care of the RBAC settings. The Service Binding Usage Controller and Console Backend Service must be allowed to perform needed operations on the resources, with the type defined in the UsageKind object.

See the example of the RBAC Rule for the Binding Usage Controller:
```yaml
- apiGroups: ["kubeless.io"]
  resources: ["functions"]
  verbs: ["get", "update"]
```
Here is the example for the Console Backend Service:
```yaml
- apiGroups: ["kubeless.io"]
  resources: ["functions"]
  verbs: ["list"]
```
