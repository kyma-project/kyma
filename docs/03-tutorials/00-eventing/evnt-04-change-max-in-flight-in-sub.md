---
title: Changing Events Max-In-Flight in Subscriptions
---

In this tutorial, you learn how to set idle "in-flight messages" limit in Kyma Subscriptions.
The "in-flight messages" config defines the number of events that Kyma Eventing forwards in parallel to the sink, without waiting for a response. 

## Prerequisites

>**NOTE:** Read about [Istio sidecars in Kyma and why you want them](https://kyma-project.io/#/istio/user/00-overview/00-30-overview-istio-sidecars). Then, check how to [enable automatic Istio sidecar proxy injection](https://kyma-project.io/#/istio/user/02-operation-guides/operations/02-20-enable-sidecar-injection). For more details, see [Default Istio setup in Kyma](https://kyma-project.io/#/istio/user/00-overview/00-40-overview-istio-setup).

1. Follow the [Prerequisites steps](./) for the Eventing tutorials.
2. [Create a Function](../../02-get-started/04-trigger-workload-with-event.md#create-a-function).
3. For this tutorial, instead of the default code sample, replace the Function source with the following code. To simulate prolonged event processing, the Function waits for 5 seconds before returning the response.

<!-- tabs:start -->

#### **Kyma Dashboard**

   ```js
   module.exports = {
     main: async function (event, context) {
       console.log("Processing event:", event.data);
       // sleep/wait for 5 seconds
       await new Promise(r => setTimeout(r, 5 * 1000));
       console.log("Completely processed event:", event.data);
       return;
     } 
   }
   ```

#### **kubectl**

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
            console.log("Processing event:", event.data);
            // sleep/wait for 5 seconds
            await new Promise(r => setTimeout(r, 5 * 1000));
            console.log("Completely processed event:", event.data);
            return;
          }
        }
EOF
```

<!-- tabs:end -->

## Create a Subscription with Max-In-Flight config

Create a [Subscription](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) custom resource. Subscribe for events of the type: `order.received.v1` and set the `maxInFlightMessages` to `5`, so that Kyma Eventing forwards maximum 5 events in parallel to the sink without waiting for a response.

<!-- tabs:start -->

#### **Kyma Dashboard**

1. Go to **Namespaces** and select the default Namespace.
2. Go to **Configuration** > **Subscriptions** and click **Create Subscription+**.
3. Switch to the **Advanced** tab, and provide the following parameters:
   - **Subscription name**: `lastorder-sub`
   - **Config**: `maxInFlightMessages: 5`
   - **Types**: `order.received.v1`
   - **Service**: `lastorder` (The sink field will be populated automatically.)
   - **Type matching:**: `standard`
   - **Source**: `myapp`
     
5. Click **Create**.
6. Wait a few seconds for the Subscription to have status `READY`.

#### **kubectl**

Run:
```bash
cat <<EOF | kubectl apply -f -
   apiVersion: eventing.kyma-project.io/v1alpha2
   kind: Subscription
   metadata:
     name: lastorder-sub
     namespace: default
   spec:
     config:
       maxInFlightMessages: "5"
     sink: 'http://lastorder.default.svc.cluster.local'
     source: myapp
     types:
       - order.received.v1
EOF
```

To check that the Subscription was created and is ready, run:
```bash
kubectl get subscriptions lastorder-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the returned status says `true`.

<!-- tabs:end -->

## Trigger the workload with multiple events

You created the `lastorder` Function, and subscribed to the `order.received.v1` events by creating a Subscription CR.
Next, publish 15 events at once and see how Kyma Eventing triggers the workload.

1. Port-forward the [Event Publisher Proxy](../../05-technical-reference/00-architecture/evnt-01-architecture.md) Service to localhost, using port `3000`. Run:
   ```bash
   kubectl -n kyma-system port-forward service/eventing-event-publisher-proxy 3000:80
   ```
2. Now publish 15 events to the Event Publisher Proxy Service. In another terminal window, run:

<!-- tabs:start -->

#### **CloudEvents Conformance Tool**

   ```bash
   for i in {1..15}
   do
     cloudevents send http://localhost:3000/publish \
       --type order.received.v1 \
       --id e4bcc616-c3a9-4840-9321-763aa23851f${i} \
       --source myapp \
       --datacontenttype application/json \
       --data "{\"orderCode\":\"$i\"}" \
       --yaml
   done
   ```

#### **curl**

   ```bash
   for i in {1..15}
   do
     curl -v -X POST \
       -H "ce-specversion: 1.0" \
       -H "ce-type: order.received.v1" \
       -H "ce-source: myapp" \
       -H "ce-eventtypeversion: v1" \
       -H "ce-id: e4bcc616-c3a9-4840-9321-763aa23851f${i}" \
       -H "content-type: application/json" \
       -d "{\"orderCode\":\"$i\"}" \
       http://localhost:3000/publish
   done
   ```

<!-- tabs:end -->

## Verify the event delivery

To verify that the events ware properly delivered, check the logs of the Function (see [Verify the event delivery](../../02-get-started/04-trigger-workload-with-event.md#verify-the-event-delivery)).

You will see the received events in the logs as:
```
Processing event: { orderCode: '1' }
Processing event: { orderCode: '2' }
Processing event: { orderCode: '3' }
Processing event: { orderCode: '4' }
Processing event: { orderCode: '5' }
Completely processed event: { orderCode: '1' }
Processing event: { orderCode: '6' }
Completely processed event: { orderCode: '2' }
Processing event: { orderCode: '7' }
Completely processed event: { orderCode: '3' }
Processing event: { orderCode: '8' }
Completely processed event: { orderCode: '4' }
Processing event: { orderCode: '9' }
Completely processed event: { orderCode: '5' }
Processing event: { orderCode: '10' }
Completely processed event: { orderCode: '6' }
Processing event: { orderCode: '11' }
Completely processed event: { orderCode: '7' }
Processing event: { orderCode: '12' }
Completely processed event: { orderCode: '8' }
Processing event: { orderCode: '13' }
Completely processed event: { orderCode: '9' }
Processing event: { orderCode: '14' }
Completely processed event: { orderCode: '10' }
Processing event: { orderCode: '15' }
Completely processed event: { orderCode: '11' }
Completely processed event: { orderCode: '12' }
Completely processed event: { orderCode: '13' }
Completely processed event: { orderCode: '14' }
Completely processed event: { orderCode: '15' }
```

You can see that only 5 events at maximum were delivered to the Function in parallel and as soon as the Function completes the processing of the event and returns the response, Kyma Eventing delivers the next in-line event to the Function. 
