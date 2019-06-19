---
title: ServiceBindingUsage
type: Custom Resource
---

The `servicebindingusages.servicecatalog.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to inject Secrets to an application. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd servicebindingusages.servicecatalog.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource in which the ServiceBindingUsage injects a Secret associated with the `redis-instance-binding` ServiceBinding to the `redis-client` Deployment in the `production` Namespace. In this example, the **status.conditions.status** field is set to `True`, which means that the ServiceBinding injection is successful. If the injection fails, this field is set to `False` and the **message** and **reason** fields appear. This example has also the **envPrefix.name** field specified, which adds a prefix to all environment variables injected from a given Secret to your Pod. This allows you to separate environment variables injected from different Secrets. By default, the prefixing is disabled. Set the **envPrefix.name** to enable it.

>**NOTE:** The prefix itself is not separated from the name of an environment variable by any character. If you want to separate your prefix, add a special character at the end of it.

```
apiVersion: servicecatalog.kyma-project.io/v1alpha1
kind: ServiceBindingUsage
metadata:
 name: redis-client-binding-usage
 namespace: production
 "ownerReferences": [
    {
       "apiVersion": "servicecatalog.k8s.io/v1beta1",
       "kind": "ServiceBinding",
       "name": "redis-instance-binding",
       "uid": "65cc140a-db6a-11e8-abe7-0242ac110023"
    }
 ],
spec:
 serviceBindingRef:
   name: redis-instance-binding
 usedBy:
   kind: deployment
   name: redis-client
 parameters:
   envPrefix:
     name: "pico-bello"
status:
    conditions:
    - lastTransitionTime: 2018-06-26T10:52:05Z
      lastUpdateTime: 2018-06-26T10:52:05Z
      status: "True"
      type: Ready
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** |    YES   | Specifies the name of the CR. |
| **metadata.namespace** |    YES   | Specifies the Namespace in which the CR is created. |
| **metadata.ownerReferences** |    YES   | Contains an ownerReference to the binding specified in the **spec.serviceBindingRef.name** field, if the binding exists. |
| **spec.serviceBindingRef.name** |    YES   | Specifies the name of the ServiceBinding. |
| **spec.usedBy** |    YES   | Specifies the application into which the Secret is injected. |
| **spec.usedBy.kind** |    YES   | Specifies the name of the UsageKind custom resource. |
| **spec.usedBy.name** |    YES   | Specifies the name of the application. |
| **spec.parameters.envPrefix** |    NO   | Adds a prefix to environment variables injected from the given Secret. The prefixing is disabled by default. |
| **spec.parameters.envPrefix.name** |    YES   | Defines the name of the prefix. This field is mandatory if **envPrefix** is specified.  |
| **status.conditions** |    NO   | Defines the state of the ServiceBindingUsage.|
| **status.conditions.lastTransitionTime** |    NO   | Specifies the first time when the Service Binding Usage Controller processed the ServiceBindingUsage, or the time when the **status.conditions.status** field changed. |
| **status.conditions.lastUpdateTime** |    NO   | Specifies the time of the last ServiceBindingUsage condition update. |
| **status.conditions.status** |    NO   |  Specifies whether the ServiceBinding injection is successful or not. |
| **status.conditions.type** |    NO   | Defines the type of the condition. The value of this field is always `ready`. |
| **message** |    NO   | Describes in a human-readable way why the ServiceBinding injection failed. |
| **reason** |    NO   | Specifies a unique, one-word, CamelCase reason for the ServiceBinding injection failure. See [this](././components/service-binding-usage-controller/internal/controller/status/usage.go) file for the complete list of reasons. |


## Related resources and components

These are the resources related to this CR:

| Custom resource   |   Description |
|----------|------|
| [UsageKind](#custom-resource-usagekind) |  Provides information on where to inject Secrets. |
| [ServiceBinding](https://kubernetes.io/docs/concepts/extend-kubernetes/service-catalog/#api-resources) |  Provides Secrets to inject.  |


These components use this CR:

| Component   |   Description |
|----------|------|
| Service Binding Usage Controller |  Reacts to every action of creating, updating, or deleting ServiceBindingUsages in all Namespaces, and uses ServiceBindingUsage data to inject binding. |
| Console Backend Service |  Exposes the given CR to the Console UI. It also allows you to create and delete a ServiceBindingUsage. |
