---
title: Event subscription
type: Custom Resource
---

The Event subscription CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to subscribe to events {more detailed description}. To get the up-to-date CRD and show the output in the yaml format, run this command:

`kubectl get crd {CRD name} -o yaml`

## Sample custom resource

The following event subscription resource subscribes to an event called `sap.kyma.custom.commerce.order.created.v1`.

> **NOTE:** Both the subscriber and the subscription should exist in the same Namespace.

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

The `sink` field refers to the cluster local name of the subscriber's service.

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   | Required |  Description |
|-------------|:---------:|--------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |
| **{another_parameter}** | {Yes/No} | {Parameter description} |



