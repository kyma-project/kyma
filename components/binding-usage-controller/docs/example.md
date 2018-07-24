# Example

In this example, the ServiceBindingUsage injects a Secret associated with the `redis-instance-binding` ServiceBinding to the `redis-client` Deployment in the `production` Namespace.


```yaml
apiVersion: servicecatalog.kyma.cx/v1alpha1
kind: ServiceBindingUsage
metadata:
 name: redis-client-binding-usage
 namespace: production
 # Objects under the spec parameter do not have a Namespace property. It indicates that all of them should be available in the same Namespace as the ServiceBindingUsage. The ServiceBinding works in the same way in the Service Catalog.
spec:
 # serviceBindingRef is the reference to the ServiceBinding and it needs to be in the same Namespace where the ServiceBindingUsage is created.
 serviceBindingRef:
   name: redis-instance-binding
 # usedBy is the reference to the application to which the Binding Usage Controller injects environment variables included in the ServiceBinding pointed by the serviceBindingRef. The pointed resource should be available in the same Namespace as the ServiceBindingUsage. The supported kinds in the usedBy section are `Development` and `Function`.
 usedBy:
   kind: Deployment
   name: redis-client
 # parameters is a set of the parameters passed to the controller
 parameters:
   # envPrefix defines prefixing of environment variables injected by the ServiceBindingUsage. This field is not required as prefixing is disabled by default.
   envPrefix:
     # name is a required field if envPrefix is specified.
     name: "pico-bello"
# status contains each action passed by the ServiceBindingUsage.
status:
    # conditions represent the observations of the ServiceBindingUsage state.
    conditions:
      # lastTransitionTime is set when the Binding Usage Controller processes the ServiceBindingUsage for the first time or when the status field changes.
    - lastTransitionTime: 2018-06-26T10:52:05Z
      # lastUpdateTime is set on each condition update. The condition is updated every time when you process the ServiceBindingUsage.
      lastUpdateTime: 2018-06-26T10:52:05Z
      # status is a boolean which determines if the ServiceBinding injection is successful.
      status: "True"
      # type defines if the condition is `ready`.
      type: Ready
```

**Conditions** can also have their **status** parameter set to `false`, in which case the **message** and **reason** fields appear. See the following example:

```yaml
- lastTransitionTime: 2018-06-22T17:27:17Z
lastUpdateTime: 2018-06-22T17:27:22Z
# message describes the state of the ServiceBindingUsage.
message: 'while getting ServiceBinding "redis-instance-credential" from namespace
  "default": servicebinding.servicecatalog.k8s.io "redis-instance-credential"
  not found'
# reason briefly describes the state of the ServiceBindingUsage.
reason: ServiceBindingGetError
status: "False"
type: Ready
```

Find the list of all **conditions** and their descriptions in [this](../internal/controller/status/usage.go) file.
For the ready-to-use examples, see the [`examples`](../examples) folder.
