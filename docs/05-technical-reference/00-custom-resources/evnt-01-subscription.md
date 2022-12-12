---
title: Subscription
type: Custom Resource
---

The `subscriptions.eventing.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to subscribe to events. To get the up-to-date CRD and show the output in the YAML format, run this command:

`kubectl get crd subscriptions.eventing.kyma-project.io -o yaml`

## Sample custom resource

This sample Subscription custom resource (CR) subscribes to an event called `order.created.v1`.

> **WARNING:** Prohibited characters in event names under the **spec.types** property, are not supported in some backends. If any are detected, Eventing may remove them. Read [Event names](../evnt-01-event-names.md#event-name-cleanup) for more information.

> **NOTE:** Both the subscriber and the Subscription should exist in the same Namespace.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha2
kind: Subscription
metadata:
  name: test
  namespace: test
spec:
  typeMatching: standard
  source: commerce
  types:
    - order.created.v1
  sink: http://test.test.svc.cluster.local
  config:
    maxInFlightMessages: 10
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   | Required |  Description |
|-------------|:---------:|--------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | No | Defines the Namespace in which the CR is available. It is set to `default` unless your specify otherwise. |
| **spec.typeMatching** | No | Defines the matching type (`standard` or `exact`) for event types. When it is set to `exact`, Eventing will not do any kind of modifications to the provided `spec.types` internally. It is set to `standard` unless you specify otherwise. |
| **spec.source** | Yes | The origin from which events are published. |
| **spec.types** | Yes | Defines the list of event types used to trigger workloads. |
| **spec.sink** | Yes | Specifies the HTTP endpoint where matching events should be sent to, for example: `test.test.svc.cluster.local`. |
| **spec.config.maxInFlightMessages** | No | The maximum idle "in-flight messages" sent by NATS to the sink without waiting for a response. By default, it is set to 10. |

## Related resources and components

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| [Eventing Controller](../00-architecture/evnt-01-architecture.md#eventing-controller) | The Eventing Controller reconciles on Subscriptions and creates a connection between subscribers and the Eventing backend. |
| [Event Publisher Proxy](../00-architecture/evnt-01-architecture.md#event-publisher-proxy) | The Event Publisher Proxy reads the Subscriptions to find out how events are used for each Application. |
