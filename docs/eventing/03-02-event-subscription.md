---
title: Event subscription
type: Details
---

When an event is created, the event subscription contains certain filters and fields that are required by Kyma. If you want to subscribe to an event, you need to create a subscription that matches this format:

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

The `sink` field refers to the Subscriber's service cluster local name.

> **NOTE:** Both the subscriber and the subscription should exist in the same Namespace.