---
title: EventingBackend
---

The EventingBackend custom resource definition (CRD) is used to know the current status of the Eventing backend. To get the up-to-date CRD and show the output in the YAML format, run this command:

```shell
kubectl get crd eventingbackends.eventing.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample custom resource (CR) that the Eventing chart creates by default when Kyma is deployed. It has an empty `spec` section.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: EventingBackend
metadata:
  name: eventing-backend
  namespace: kyma-system
spec: {}
```

## Additional information

When you fetch an existing EventingBackend CR, the system adds the **status** section which describes the status of the Eventing Controller and the Eventing Publisher Proxy. This table lists the fields of the **status** section.

| Field   |  Description |
|:---|:---|
| **status.backendType** | Specifies the backend type used. |
| **status.conditions.code** | Conditions defines the ready status of the Eventing Controller and the Eventing Publisher Proxy. |
| **status.eventingReady** | Current ready status of the Eventing Backend. |

## Related resources and components

These components use this CR:

| Component           | Description                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| [Eventing Controller](../00-architecture/evnt-01-architecture.md#eventing-controller) | The Eventing Controller uses this CR to display its ready status and the ready status of the [Event Publisher Proxy](../00-architecture/evnt-01-architecture.md#event-publisher-proxy). |

