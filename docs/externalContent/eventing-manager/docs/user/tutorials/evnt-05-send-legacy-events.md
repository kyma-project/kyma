# Publish Legacy Events Using Kyma Eventing

Kyma Eventing also supports sending and receiving of legacy events. In this tutorial we will show how to send legacy events.

> [!NOTE]
> It is recommended to use [CloudEvents specification](https://cloudevents.io/) for sending and receiving events in Kyma.

## Prerequisites

> [!NOTE]
> Read about the [Purpose and Benefits of Istio Sidecar Proxies](https://kyma-project.io/#/istio/user/00-00-istio-sidecar-proxies?id=purpose-and-benefits-of-istio-sidecar-proxies). Then, check how to [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection). For more details, see [Default Istio Configuration](https://kyma-project.io/#/istio/user/00-15-overview-istio-setup) in Kyma.

1. Follow the [Prerequisites steps](evnt-01-prerequisites.md) for the Eventing tutorials.
2. [Create and Modify an Inline Function](https://kyma-project.io/#/serverless-manager/user/tutorials/01-10-create-inline-function).

## Create a Subscription

To subscribe to events, we need a [Subscription](../resources/evnt-cr-subscription.md) custom resource (CR). We're going to subscribe to events of the type `order.received.v1`.

<Tabs>
<Tab name="Kyma Dashboard">

1. Go to **Namespaces** and select the default namespace.
2. Go to **Configuration** > **Subscriptions** and click **Create Subscription+**.
3. Provide the following parameters:
   - **Subscription name**: `lastorder-sub`
   - **Types**: `order.received.v1`
   - **Service**: `lastorder` (The sink field will be populated automatically.)
   - **Type matching:**: `standard`
   - **Source**: `myapp`

4. Click **Create**.
5. Wait a few seconds for the Subscription to have status `READY`.
</Tab>
<Tab name="curl">

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
</Tab>
</Tabs>

## Publish a Legacy Event To Trigger the Workload

You created the `lastorder` Function, and subscribed to the `order.received.v1` events by creating a Subscription CR. Now it's time to send an event and trigger the Function.

1. Port-forward the [Eventing Publisher Proxy](../evnt-architecture.md) Service to localhost, using port `3000`. Run:

   ```bash
   kubectl -n kyma-system port-forward service/eventing-publisher-proxy 3000:80
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

> [!NOTE]
> If you want to use a Function to publish a CloudEvent, see the [Event object SDK specification](https://kyma-project.io/#/serverless-manager/user/technical-reference/07-70-function-specification?id=event-object-sdk).

## Publish Legacy Events

To verify that the event was properly delivered, check the logs of the Function:

<Tabs>
<Tab name="Kyma Dashboard">

1. In Kyma Dashboard, return to the view of your `lastorder` Function.
2. In the **Code** view, find the **Replicas of the Function** section.
3. Click the name of your replica.
4. Locate the **Containers** section and click on **View Logs**.
</Tab>
<Tab name="kubectl">

Run:

```bash
kubectl logs \
  -n default \
  -l serverless.kyma-project.io/function-name=lastorder,serverless.kyma-project.io/resource=deployment \
  -c function
```
</Tab>
</Tabs>

You see the received event in the logs:

```sh
Received event: { orderCode: '3211213' }
```
