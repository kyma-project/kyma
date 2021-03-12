---
title: Event subscription
type: Details
---

When an event is created, the event subscription contains certain filters and fields that are required by Kyma. If you want to subscribe to an event, you need to create a subscription that matches this format:

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: Subscription
metadata:
  creationTimestamp: "2021-03-04T07:15:20Z"
  finalizers:
  - eventing.kyma-project.io
  generation: 1
  labels:
    eventing.knative.dev/broker: default
  name: vibrant-hellman
  namespace: test
  ownerReferences:
  - apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    name: test
    uid: f982d53d-02c2-40ef-ba0b-78373b61bfe3
  resourceVersion: "33672"
  selfLink: /apis/eventing.kyma-project.io/v1alpha1/namespaces/test/subscriptions/vibrant-hellman
  uid: becce239-15c0-469f-9529-ade2883e003c
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
  sink: http://test.test.svc.cluster.local:80/
```