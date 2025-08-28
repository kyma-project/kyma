# Event Name Cleanup in Subscriptions

To conform to Cloud Event specifications, sometimes Eventing must modify the event names to filter out prohibited characters. This tutorial presents one example of event name cleanup.
You learn how Eventing behaves when you create a [Subscription](../resources/evnt-cr-subscription.md) having prohibited characters in the event names. Read more about [Event name format and cleanup](../evnt-event-names.md).

## Prerequisites

> [!NOTE]
> Read about the [Purpose and Benefits of Istio Sidecar Proxies](https://kyma-project.io/#/istio/user/00-00-istio-sidecar-proxies?id=purpose-and-benefits-of-istio-sidecar-proxies). Then, check how to [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection). For more details, see [Default Istio Configuration](https://kyma-project.io/#/istio/user/00-15-overview-istio-setup) in Kyma.

1. Follow the [Prerequisites steps](evnt-01-prerequisites.md) for the Eventing tutorials.
2. [Create and Modify an Inline Function](https://kyma-project.io/#/serverless-manager/user/tutorials/01-10-create-inline-function).
3. For this tutorial, instead of the default code sample, replace the Function source with the following code:

<!-- tabs:start -->

#### Kyma Dashboard

```js
module.exports = {
  main: async function (event, context) {
    console.log("Received event: ", event.data, ", Event Type: ", event.extensions.request.headers['ce-type']);
    return;
  } 
}
```

#### kubectl

```bash
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: lastorder
  namespace: default
spec:
  replicas: 1
  resourceConfiguration:
    function:
      profile: S
    build:
      profile: local-dev
  runtime: nodejs18
  source:
    inline:
      source: |-
        module.exports = {
          main: async function (event, context) {
            console.log("Received event: ", event.data, ", Event Type: ", event.extensions.request.headers['ce-type']);
            return;
         }
       }
   EOF
```

<!-- tabs:end -->

## Create a Subscription With Event Type Consisting of Alphanumeric Characters

Create a [Subscription](../resources/evnt-cr-subscription.md) custom resource (CR) and subscribe for events of the type: `order.payment*success.v1`. Note that `order.payment*success.v1` contains a prohibited character, the asterisk `*`.

<!-- tabs:start -->

#### Kyma Dashboard

1. Go to **Namespaces** and select the default namespace.
2. Go to **Configuration** > **Subscriptions** and click **Create Subscription+**.
3. Provide the following parameters:
   - **Subscription name**: `lastorder-payment-sub`
   - **Types**: `order.payment*success.v1`
   - **Service**: `lastorder` (The sink field will be populated automatically.)
   - **Type matching:**: `standard`
   - **Source**: `myapp`

4. Click **Create**.
5. Wait a few seconds for the Subscription to have status `READY`.

#### kubectl

Run:

```bash
cat <<EOF | kubectl apply -f -
   apiVersion: eventing.kyma-project.io/v1alpha2
   kind: Subscription
   metadata:
     name: lastorder-payment-sub
     namespace: default
   spec:
     sink: 'http://lastorder.default.svc.cluster.local'
     source: myapp
     types:
       - order.payment*success.v1
EOF
```

To check that the Subscription was created and is ready, run:

```bash
kubectl get subscriptions lastorder-payment-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the returned status says `true`.

<!-- tabs:end -->

## Check the Subscription Cleaned Event Type

To check the Subscription cleaned Event type, run:

```bash
kubectl get subscriptions lastorder-payment-sub -o=jsonpath="{.status.types}"
```

Note that the returned event type `["order.paymentsuccess.v1"]` does not contain the asterisk `*` in the `payment*success` part. That's because Kyma Eventing cleans out the prohibited characters from the event name and uses the cleaned event name in the underlying Eventing backend.

## Trigger the Workload With an Event

You created the `lastorder` Function, and subscribed to the `order.payment*success.v1` events by creating a Subscription CR.
Next, you see that you can still publish events with the original Event name (i.e. `order.payment*success.v1`) even though it contains the prohibited character, and it triggers the Function.

1. Port-forward the [Eventing Publisher Proxy](../evnt-architecture.md) Service to localhost, using port `3000`. Run:

   ```bash
   kubectl -n kyma-system port-forward service/eventing-publisher-proxy 3000:80
   ```

2. Publish an event to trigger your Function. In another terminal window, run:

<!-- tabs:start -->

#### CloudEvents Conformance Tool

```bash
cloudevents send http://localhost:3000/publish \
   --type "order.payment*success.v1" \
   --id e4bcc616-c3a9-4840-9321-763aa23851fc \
   --source myapp \
   --datacontenttype application/json \
   --data "{\"orderCode\":\"3211213\", \"orderAmount\":\"1250\"}" \
   --yaml
```

#### curl

```bash
curl -v -X POST \
     -H "ce-specversion: 1.0" \
     -H "ce-type: order.payment*success.v1" \
     -H "ce-source: myapp" \
     -H "ce-eventtypeversion: v1" \
     -H "ce-id: e4bcc616-c3a9-4840-9321-763aa23851fc" \
     -H "content-type: application/json" \
     -d "{\"orderCode\":\"3211213\", \"orderAmount\":\"1250\"}" \
     http://localhost:3000/publish
```

<!-- tabs:end -->

## Verify the Event Delivery

To verify that the event was properly delivered, check the logs of the Function:

<!-- tabs:start -->

#### Kyma Dashboard

1. In Kyma Dashboard, return to the view of your `lastorder` Function.
2. In the **Code** view, find the **Replicas of the Function** section.
3. Click the name of your replica.
4. Locate the **Containers** section and click on **View Logs**.

#### kubectl

Run:

```bash
kubectl logs \
  -n default \
  -l serverless.kyma-project.io/function-name=lastorder,serverless.kyma-project.io/resource=deployment \
  -c function
```

<!-- tabs:end -->

You see the received event in the logs:

```sh
Received event:  { orderCode: '3211213', orderAmount: '1250' } , Event Type:  order.paymentsuccess.v1
```

Note that the `Event Type` of the received event is not the same as defined in the Subscription.

## Conclusion

You see that Kyma Eventing modifies the event names to filter out prohibited characters to conform to Cloud Event specifications.

> [!WARNING]
> This cleanup modification is abstract; you can still publish and subscribe to the original Event names. However, in some cases, it can lead to a naming collision as explained in [Event names](../evnt-event-names.md).
