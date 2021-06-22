---
title: Send events without a Kyma Application
type: Tutorials
---

**In-cluster Eventing** allows publishers to send messages and subscribers to receive them without the need for a Kyma Application. This means that instead of the usual event flow where Application Connector publishes events to the Event Publisher Proxy, events can be published from within the cluster directly to the Event Publisher Proxy.


## Prerequisites

- A running Kyma cluster with a valid Namespace

## Steps

1. In Kyma's left navigation panel, go to **Workloads** > **Functions** and navigate to your Function.

2. Once in the Function details view, switch to the **Configuration** tab, and select **Create Event Subscription** in the **Event Subscriptions** section.

3. Create a subscription where the **eventType.value** field includes the name of your Application. In this example, this is `sap.kyma.custom.nonexistingapp.order.created.v1`, where `nonexistingapp` is an Application that does not exist in Kyma.

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

```


4. Select the event type and version that you want to use for your Function and select **Save** to confirm changes.

You get a confirmation that the Event Subscription was successfully created, and you will see it in the **Event Subscriptions** section in your Function.


5. On the event publisher's side, include the exact same Application name in the `type` field, like in this example:

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
