---
title: ServiceBindingUsage
type: Custom Resource
---

The `servicebindingusages.servicecatalog.kyma.cx` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to inject Secrets to the application. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd servicebindingusages.servicecatalog.kyma.cx -o yaml
```

## Sample custom resource

This is a sample resource in which the ServiceBindingUsage injects a Secret associated with the `redis-instance-binding` ServiceBinding to the `redis-client` Deployment in the `production` Namespace. This example has the **conditions.status** field set to `true`, which means that the ServiceBinding injection is successful. If this field is set to `false`, the **message** and **reason** fields appear.

```
apiVersion: servicecatalog.kyma.cx/v1alpha1
kind: ServiceBindingUsage
metadata:
 name: redis-client-binding-usage
 namespace: production
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


| Parameter   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **metadata.namespace** |    **YES**   | Specifies the Namespace in which the CR is created. |
| **spec.serviceBindingRef.name** |    **YES**   | Specifies the name of the ServiceBinding. |
| **spec.usedBy** |    **YES**   | Specifies the application into which the Secret is injected. |
| **spec.usedBy.kind** |    **YES**   | Specifies the name of the [UsageKind](041-cr-usage-kind.md). |
| **spec.usedBy.name** |    **YES**   | Specifies the name of the application. |
| **spec.parameters.envPrefix** |    **NO**   | Defines the prefix of environment variables environment variables that the ServiceBindingUsage injects. The prefixing is disabled by default. |
| **spec.parameters.envPrefix.name** |    **YES**   | Specifies the name of the prefix. This field is mandatory if **envPrefix** is specified.  |
| **status.conditions** |    **NO**   | Specifies the state of the ServiceBindingUsage.|
| **status.conditions.lastTransitionTime** |    **NO**   | Specifies the time when the Binding Usage Controller processes the ServiceBindingUsage for the first time or when the **status.conditions.status** field changes. |
| **status.conditions.lastUpdateTime** |    **NO**   | Specifies the time of the last ServiceBindingUsage condition update. |
| **status.conditions.status** |    **NO**   |  Specifies whether the status of the **status.conditions.type** field is `True` or `False`. |
| **status.conditions.type** |    **NO**   | Defines the type of the condition. The value of this field is always `Ready`. |
| **message** |    **NO**   | Describes in a human-readable way why the ServiceBinding injection has failed. |
| **reason** |    **NO**   | Specifies a unique, one-word, CamelCase reason for the condition's last transition. |
