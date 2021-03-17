---
title: Subscription
type: Custom Resource
---

The Subscription CustomResourceDefinition (CRD) is used to subscribe to events. To get the up-to-date CRD and show the output in the yaml format, run this command:

`kubectl get crd subscriptions.eventing.kyma-project.io -o yaml`

## Sample custom resource

This sample Subscription resource subscribes to an event called `sap.kyma.custom.commerce.order.created.v1`.

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
| **metadata.namespace** | No | Defines the Namespace in which the CR is available. It is set to `default` unless your specify otherwise. |
| **spec.filter** | Yes | Defines the list of filters. |
| **spec.filter.dialect** | No | Specifies the preferred Eventing backend. Currently, the capability to switch between Eventing backends is not available. It is set to NATS by default. |
| **spec.filter.filters** | Yes | Defines the filter element as a combination of two Cloud Event filter elements. |
| **spec.filter.filters.eventSource** | Yes | The origin from which events are published. |
| **spec.filter.filters.eventType** | Yes | The type of events used to trigger workloads. |
| **spec.filter.filters.eventSource.property** | Yes | Must be set to `source`. |
| **spec.filter.filters.eventSource.type** | No | Must be set to `exact`. |
| **spec.filter.filters.eventSource.value** | Yes | Must be set to `""` for the NATS backend. |
| **spec.filter.filters.eventType.property** | Yes | Must be set to `type`. |
| **spec.filter.filters.eventType.type** | No | Must be set to `exact`. |
| **spec.filter.filters.eventType.value** | Yes | Name of the event being subscribed to, for example: `sap.kyma.custom.commerce.order.created.v1`. |
| **spec.protocol** | Yes | Must be set to `""`. |
| **spec.protocolsettings** | Yes | Defines the Cloud Event protocol setting specification implementation. Must be set to `{}`. |
| **spec.sink** | Yes | Specifies the HTTP endpoint where matching events should be sent to, for example: `test.test.svc.cluster.local`.  |

## Related resources and components

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| Eventing Controller | The Eventing Controller reconciles on Subscriptions and creates a connection between subscribers and the Eventing backend. |
| Event Publisher Proxy | The Event Publisher Proxy reads the Subscriptions to find out how events are used for each Application. |