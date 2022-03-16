# Binding Usage Target proposal

## Overview

Usage Kind allows dynamic, in the runtime, Binding Usage Controller (BUC) configuration. It allows you to define what resources can be bound with Service Binding and how to bind them. Currently, you can bind only Deployment and Function resources. Adding a new resource requires changes in the Binding Usage Controller code. The goal is to make it possible only with the configuration.

## Usage Kind 

The idea is to have a new cluster-wide custom resource which defines a bindable resource.

See the example:

```yaml
apiVersion: servicecatalog.kyma-project.io/v1alpha1
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

## Console Backend Service 

UI needs two new endpoints which return:
 * all kinds which you can use with the binding usage. It is a list of all usage kinds.
 * all resources with a given kind. The result must be filtered out by the **metadata.ownerReference** field. If the resource contains **metadata.ownerReference**, the user should not see such a resource in the UI.
## Binding Usage Controller

The Binding Usage Controller takes the value of the **spec.usedBy.kind** field of the Service Binding Usage resource, looks for the corresponding Usage Kind which contains information about the kind of resource to be bound. The controller must handle resources even if the specified **labelsPath** field does not exist.

## Security

The administrator who adds Usage Kind must take care of RBAC settings. BUC must be allowed to perform needed operations on resources with the kind defined in the Usage Kind resource.

## Example

There is a defined Usage Kind:
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
There is also a RBAC role with the following rule for the Binding Usage Controller:
```yaml
- apiGroups: ["kubeless.io"]
  resources: ["functions"]
  verbs: ["get", "update"]
```
and for the Console Backend Service:
```yaml
- apiGroups: ["kubeless.io"]
  resources: ["functions"]
  verbs: ["list"]
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
apiVersion: servicecatalog.kyma-project.io/v1alpha1
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
The Service Binding Usage **spec.usedBy.kind** field matches the name of the Usage Kind instance.

## Testing

It is important to have an easy way to create a test for a new Usage Kind. There should be a way to check that the proper RBAC role is set or that the **labelsPath** field is correct. The purpose is to design a solution that makes writing such acceptance tests as easy as possible.

## Usage Kind and Service Binding Usage finalizers

The Binding Usage Controller implementation cannot handle the Service Binding Usage deletion if the Usage Kind has been deleted. The idea is to have a finalizer in the Usage Kind instance. The finalizer is removed only when the Usage Kind is not used by any Service Binding Usage. The Binding Usage Controller requires the Usage Kind to delete the Service Binding Usage. For this reason, a finalizer is needed in the Service Binding Usage. The deletion scenario looks as follows:

1. In this scenario, there are the following Usage Kind and Service Binding Usage:
```yaml
apiVersion: servicecatalog.kyma-project.io/v1alpha1
kind: UsageKind
metadata:
  clusterName: ""
  creationTimestamp: 2018-08-06T08:02:19Z
  finalizers:
  - servicecatalog.kyma-project.io/usage-kind-protection
  generation: 1
  name: deployment
  namespace: ""
  resourceVersion: "10984"
  selfLink: /apis/servicecatalog.kyma-project.io/v1alpha1/usagekinds/deployment
  uid: 0a7716ef-994f-11e8-98a9-560fb490844b
spec:
  displayName: Deployment
  labelsPath: spec.template.metadata.labels
  resource:
    group: apps
    kind: deployment
    version: v1beta1
```

```yaml
apiVersion: servicecatalog.kyma-project.io/v1alpha1
  kind: ServiceBindingUsage
  metadata:
    clusterName: ""
    creationTimestamp: 2018-08-06T08:02:34Z
    finalizers:
      - servicecatalog.kyma-project.io/sbu-protection
    generation: 1
    name: deploy-redis-client
    namespace: default
    resourceVersion: "11016"
    selfLink: /apis/servicecatalog.kyma-project.io/v1alpha1/namespaces/default/servicebindingusages/deploy-redis-client
    uid: 13c260e7-994f-11e8-98a9-560fb490844b
  spec:
    serviceBindingRef:
      name: redis-instance-credential
    usedBy:
      kind: deployment
      name: redis-client
  status:
    conditions:
    - lastTransitionTime: 2018-08-06T08:02:35Z
      lastUpdateTime: 2018-08-06T08:02:35Z
      status: "True"
      type: Ready
```

2. The administrator performs the Usage Kind deletion.
3. The Usage Kind Protection Controller handles the Usage Kind update. The controller does not remove the Usage Kind finalizer as it is used by the Service Binding Usage. The Usage Kind looks as follows:
```yaml
apiVersion: servicecatalog.kyma-project.io/v1alpha1
kind: UsageKind
metadata:
  clusterName: ""
  creationTimestamp: 2018-08-06T08:02:19Z
  deletionGracePeriodSeconds: 0
  deletionTimestamp: 2018-08-06T08:15:09Z
  finalizers:
  - servicecatalog.kyma-project.io/usage-kind-protection
  generation: 2
  name: deployment
  namespace: ""
  resourceVersion: "12019"
  selfLink: /apis/servicecatalog.kyma-project.io/v1alpha1/usagekinds/deployment
  uid: 0a7716ef-994f-11e8-98a9-560fb490844b
spec:
  displayName: Deployment
  labelsPath: spec.template.metadata.labels
  resource:
    group: apps
    kind: deployment
    version: v1beta1
``` 
4. The administrator requests deletion of the Service Binding Usage. 
5. The Binding Usage Controller handles the deletion request and removes the Service Binding Usage finalizer.
6. Kubernetes removes the Service Binding Usage instance.
7. The Usage Kind Protection Controller handles the Service Binding Usage deletion and checks, if the Usage Kind is used by any Service Binding Usage. Then, the Usage Kind Protection Controller removes the Usage Kind finalizer.
8. Kubernetes removes the Usage Kind instance.

Without Service Binding Usage finalizers the implementation is not clean. The controller, which manages Usage Kind finalizers, is triggered by the main controller instead of k8s which is not good. The cache in the indexer is not up to date (need investigation) and processing is not going through the working queue. It can cause a race condition. 

## Usage Kind snapshots in the Config Map storage

The controller can store all information about the Usage Kind used in the Service Binding Usage in the storage as a Usage Kind snapshot. Such approach allows you to modify a Usage Kind instance without losing information about the used **labelsPath** value. Handling Service Binding Usage deletion does not require existing Usage Kind. The BUC could works fine without Usage Kind finalizer. 

## Status

Proposed on 2018-07-20.

Updated on 2018-08-06.
