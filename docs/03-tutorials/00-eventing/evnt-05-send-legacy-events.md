---
title: Publish legacy events using Kyma Eventing
---

Kyma Eventing also supports sending and receiving of legacy events. In this tutorial we will show how to send legacy events.

> **NOTE:** It is recommended to use [CloudEvents specification](https://cloudevents.io/) for sending and receiving events in Kyma.

## Prerequisites

>**NOTE:** Read about [Istio sidecars in Kyma and why you want them](/istio-operator/user/00-overview/00-30-overview-istio-sidecars). Then, check how to [enable automatic Istio sidecar proxy injection](/istio-operator/user/02-operation-guides/operations/02-20-enable-sidecar-injection). For more details, see [Default Istio setup in Kyma](/istio-operator/user/00-overview/00-40-overview-istio-setup).

1. Follow the [Prerequisites steps](./) for the Eventing tutorials.
2. [Create a Function](../../02-get-started/04-trigger-workload-with-event.md#create-a-function).

## Create a Subscription

To subscribe to events, we need a [Subscription](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) custom resource. We're going to subscribe to events of the type `order.received.v1`.

<div tabs name="Create a Subscription" group="trigger-workload">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. Go to **Namespaces** and select the default Namespace.
2. Go to **Configuration** > **Subscriptions** and click **Create Subscription+**.
3. Provide the following parameters:
   - **Subscription name**: `lastorder-sub`
   - **Types**: `order.received.v1`
   - **Service**: `lastorder` (The sink field will be populated automatically.)
   - **Type matching:**: `standard`
   - **Source**: `myapp`

4. Click **Create**.
5. Wait a few seconds for the Subscription to have status `READY`.

  </details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

Run:
```bash
cat <<EOF | kubectl apply -f -
   apiVersion: eventing.kyma-project.io/v1alpha2
   kind: Subscription
   metadata:
     name: lastorder-sub
     namespace: default
   spec:
     sink: http://lastorder.default.svc.cluster.local
     source: myapp
     types:
       - order.received.v1
EOF
```

To check that the Subscription was created and is ready, run:
```bash
kubectl get subscriptions lastorder-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the command returns `true`.

  </details>
</div>

## Publish a legacy event to trigger the workload

You created the `lastorder` Function, and subscribed to the `order.received.v1` events by creating a Subscription CR. Now it's time to send an event and trigger the Function.

1. Port-forward the [Event Publisher Proxy](../../05-technical-reference/00-architecture/evnt-01-architecture.md) Service to localhost, using port `3000`. Run:
   ```bash
   kubectl -n kyma-system port-forward service/eventing-event-publisher-proxy 3000:80
   ```
2. Publish an event to trigger your Function. In another terminal window, run:

   ```bash
   curl -v -X POST \
       --data @<(<<EOF
       {
           "event-type": "order.received",
           "event-type-version": "v1",
           "event-time": "2020-09-28T14:47:16.491Z",
           "data": {"orderCode":"3211213"}
       }
   EOF
       ) \
       -H "Content-Type: application/json" \
       http://localhost:3000/myapp/v1/events
   ```

   > **NOTE:** If you want to use a Function to publish a CloudEvent, see the [Event object SDK specification](https://kyma-project.io/#/serverless-manager/user/technical-reference/07-70-function-specification?id=event-object-sdk).

## Verify the legacy event delivery

To verify that the event was properly delivered, check the logs of the Function (see [Verify the event delivery](../../02-get-started/04-trigger-workload-with-event.md#verify-the-event-delivery)).

You see the received event in the logs:
```
Received event: { orderCode: '3211213' }
```
