---
title: Subscription
type: Custom Resource
---

The Subscription CustomResourceDefinition (CRD) is used to subscribe to events and a {more detailed description}. To get the up-to-date CRD and show the output in the yaml format, run this command:

`kubectl get crd subscriptions.eventing.kyma-project.io -o yaml`

## Sample custom resource

The following Subscription resource subscribes to an event called `sap.kyma.custom.commerce.order.created.v1`.

> **NOTE:** Both the subscriber and the Subscription should exist in the same Namespace.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: Subscription
metadata:
  name: test
  namespace: test
spec:
  filter:
    filters:
    - eventSource:
        property: source
        type: exact
        value: ""
      eventType:
        property: type
        type: exact
        value: sap.kyma.custom.commerce.order.created.v1
  protocol: ""
  protocolsettings: {}
  sink: http://test.test.svc.cluster.local
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   | Required |  Description |
|-------------|:---------:|--------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | No | Defines the Namespace in which the CR is available. It is set to `default` unless your specify otherwise |
| **spec.filter** | Yes | Defines the list of filters |
| **spec.filter.dialect** | No | Specifies preferred eventing backend. Current release don't provide the capability to switch eventing backends. It is set to nats by default. |
| **spec.filter.filters** | Yes | Defines the filter element as a combination of two CE filter elements |
| **spec.filter.filters.eventSource** | Yes | Defines the event source in the eventing backend |
| **spec.filter.filters.eventType** | Yes | Defines the filter for event type |
| **spec.filter.filters.eventSource.property** | Yes | {Should be set to `source` |
| **spec.filter.filters.eventSource.type** | No | Should be set to `exact` |
| **spec.filter.filters.eventSource.value** | Yes | Can be set to "" for nats backend |
| **spec.filter.filters.eventType.property** | Yes | Should be set to `type` |
| **spec.filter.filters.eventType.type** | No | Should be set to `exact` |
| **spec.filter.filters.eventType.value** | Yes | Name of the event to be subscribed to |
| **spec.protocol** | Yes | Should be set to "" |
| **spec.protocolsettings** | Yes | Should be set to {} |
| **spec.sink** | Yes | Specifies where should matching events be sent to |

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|-----------------|---------------|
| {Related CRD kind} |  {Briefly describe the relation between the resources}. |

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| Eventing Controller |  {Briefly describe the relation between the CR and the given component}. |