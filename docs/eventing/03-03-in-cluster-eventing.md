---
title: In-cluster eventing
type: Details
---

Kyma in-cluster Eventing allows publishers to send messages and subscribers to receive them without the need for a Kyma Application. This means that instead of the usual event flow where Application Connector publishes events to the NATS Publisher Proxy, events can be published from within the cluster.

To use in-cluster Eventing, you need to create a subscription where the `eventType value` field includes the name of your application. In this example, this is `sap.kyma.custom.nonexistingapp.order.created.v1`, where `nonexistingapp` is an application that does not exist in Kyma.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: Subscription
metadata:
  name: mysub
  namespace: mynamespace
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
        value: sap.kyma.custom.nonexistingapp.order.created.v1
  protocol: ""
  protocolsettings: {}
  sink: http://myservice.mynamespace.svc.cluster.local
---
```

On the publisher side, you need to include the exact same app name in the `type:` field like in this example:

```yaml
curl -k -i \
    --data @<(cat <<EOF
    {
        "source": "kyma",
        "specversion": "1.0",
        "eventtypeversion": "v1",
        "data": {"orderCode":"3211213"},
        "datacontenttype": "application/json",
        "id": "759815c3-b142-48f2-bf18-c6502dc0998f",
        "type": "sap.kyma.custom.nonexistingapp.order.created.v1"
    }
EOF
    ) \
    -H "Content-Type: application/cloudevents+json" \
    "http://eventing-event-publisher-proxy.kyma-system/publish"
```
