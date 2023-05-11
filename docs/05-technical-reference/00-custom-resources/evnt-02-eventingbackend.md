---
title: EventingBackend
---

The `eventingbackends.eventing.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data used to manage Eventing backends within Kyma. To get the up-to-date CRD and show the output in the YAML format, run this command:

```shell
kubectl get crd eventingbackends.eventing.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample EventingBackend custom resource (CR) that the Eventing Controller creates by default when Kyma is deployed. It has an empty `spec` section.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: EventingBackend
metadata:
  name: eventing-backend
  namespace: kyma-system
spec: {}
```

## Custom resource parameters

When you fetch an existing EventingBackend CR, the Eventing Controller adds the **status** section, which shows the current status of Kyma Eventing. 

<!-- TABLE-START -->
### EventingBackend.eventing.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **backendType**  | string | Specifies the backend type used. Allowed values are "BEB" and "NATS" |
| **bebSecretName**  | string | The name of the secret containing BEB access tokens, required only for BEB |
| **bebSecretNamespace**  | string | The namespace of the secret containing BEB access tokens, required only for BEB |
| **conditions**  | \[\]object | Defines the status of the Controller and the EPP |
| **conditions.&#x200b;lastTransitionTime**  | string | Defines when the condition status was changes last |
| **conditions.&#x200b;message**  | string | Provides more detail about the condition status change |
| **conditions.&#x200b;reason**  | string | Defines why the condition status changed |
| **conditions.&#x200b;status** (required) | string |  |
| **conditions.&#x200b;type**  | string |  |
| **eventingReady**  | boolean | Defines the overall status of the Backend. |

<!-- TABLE-END -->

## Related resources and components

These components use this CR:

| Component           | Description                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| [Eventing Controller](../00-architecture/evnt-01-architecture.md#eventing-controller) | The Eventing Controller uses this CR to display its ready status and the ready status of the [Event Publisher Proxy](../00-architecture/evnt-01-architecture.md#event-publisher-proxy). |

