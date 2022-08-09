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

## Additional information

When you fetch an existing EventingBackend CR, the Eventing Controller adds the **status** section, which shows the current status of Kyma Eventing. The following table lists the fields of the **status** section.

| Field   |  Description |
|:---|:---|
| **status.backendType** | Specifies the backend type used. |
| **status.conditions.code** | Conditions defines the ready status of the Eventing Controller and the Eventing Publisher Proxy. |
| **status.eventingReady** | Current ready status of the EventingBackend. |

The `status` field of this CR looks like this:

```shell
status:
  backendType: NATS
  conditions:
  - lastTransitionTime: "2022-07-05T06:07:57Z"
    reason: Publisher proxy deployment ready
    status: "True"
    type: Publisher Proxy Ready
  - lastTransitionTime: "2022-07-05T06:07:57Z"
    reason: Subscription controller started
    status: "True"
    type: Subscription Controller Ready
  eventingReady: true
```

## Related resources and components

These components use this CR:

| Component           | Description                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| [Eventing Controller](../00-architecture/evnt-01-architecture.md#eventing-controller) | The Eventing Controller uses this CR to display its ready status and the ready status of the [Event Publisher Proxy](../00-architecture/evnt-01-architecture.md#event-publisher-proxy). |

