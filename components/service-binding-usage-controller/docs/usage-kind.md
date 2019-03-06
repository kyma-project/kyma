# UsageKind

UsageKind allows dynamic, in the runtime, Binding Usage Controller configuration. It allows you to define which resources can be bound with the ServiceBinding and how to bind them.

## UsageKind structure

From the technical point of view, the binding process modifies the target resource by adding labels to the Pod. The UsageKind specifies the field which contains labels added to the Pod, and the resource type defined by the **group**, **kind**, and **version** fields.

## Binding Usage Controller

The Binding Usage Controller takes the value from the **spec.usedBy.kind** field of the ServiceBindingUsage resource and looks for the corresponding UsageKind which contains information about the resource to be bound. The controller handles resources even if the specified **labelsPath** field does not exist.

## Finalizer

The UsageKind contains a finalizer which prevents deletion of the UsageKind in use. The controller removes the finalizer only when the UsageKind is not referenced by any ServiceBindingUsage.

## RBAC settings

The administrator who adds the UsageKind must take care of the RBAC settings. The BUC and ui-api-layer must be allowed to perform needed operations on the resources, with the type defined in the UsageKind object.

See the example of the RBAC Rule for the Binding Usage Controller:
```yaml
- apiGroups: ["kubeless.io"]
  resources: ["functions"]
  verbs: ["get", "update"]
```
Here is the example for the ui-api-layer:
```yaml
- apiGroups: ["kubeless.io"]
  resources: ["functions"]
  verbs: ["list"]
```