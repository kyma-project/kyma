# Binding Usage Target proposal

## Overview

**UsageKind** allows dynamic, in the runtime, Binding Usage Controller configuration. It allows you to define what resources can be bound with **ServiceBinding** and how to bind them. Currently, you can bind only **Deployment** and **Function** resources. Adding a new resource requires changes in the Binding Usage Controller code. The goal is to make it possible only with the configuration.

## UsageKind 

The idea is to have a new cluster-wide custom resource which defines a bindable resource.

See the example:

```yaml
apiVersion: servicecatalog.kyma.cx/v1alpha1
kind: UsageKind
metadata:
   name: function
spec:
   # displayName is a human-readable name of the usage kind
   displayName: Function
   
   # the type of target resource labeled by the controller is specified by its resource group, kind, and version. All of these fields are required.
   # there is a plan to make the version field optional in the future - the controller can find the preferred version
   resource:
     group: kubeless.io
     kind: function
     version: v1beta1
     
   # labelsPath specifies a resource field which contains labels
   labelsPath: spec.deployment.spec.template.metadata.labels
```

## Ui-api-layer

UI needs two new endpoints which return:
 * all kinds which you can use with the binding usage. It is a list of all usage kinds.
 * all resources with a given kind. The result must be filtered out by the **metadata.ownerReference** field. If the resource contains **metadata.ownerReference**, the user should not see such a resource in the UI.
## Binding Usage Controller

**Binding Usage Controller** takes the value of the `spec.usedBy.kind` field of the binding usage resource, looks for the corresponding usage kind which contains information about the kind of resource to be bound. The controller must handle resources even if the specified `labelsPath` field does not exist.

## Security

The administrator who adds **UsageKind** must take care of RBAC settings. BUC and ui-api-layer must be allowed to perform needed operations on resources with the kind defined or referenced in the usage kind object.

## Example

There is a defined **UsageKind**:
```yaml
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
There is also a RBAC role with the following rule::
```yaml
- apiGroups: ["kubeless.io"]
  resources: ["functions"]
  verbs: ["get","list","watch", "update"]
```
The user creates binding and binding usage:
```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: redis-instance-credential
spec:
  instanceRef:
    name: redis
---
apiVersion: servicecatalog.kyma.cx/v1alpha1
kind: ServiceBindingUsage
metadata:
 name: fn-redis-client
spec:
  serviceBindingRef:
    name: redis-instance-credential
  usedBy:
    kind: function
    name: redis-client
```
The ServiceBindingUsage **spec.usedBy.kind** field matches the name of the **UsageKind** instance.

## Testing

It is important to have an easy way to create a test for a new **UsageKind**. There should be a way to check that the proper RBAC role is set or that the **labelsPath** field is correct. The purpose is to design a solution that makes writing such acceptance tests as easy as possible.

## Status

Proposed at 2018.07.20
