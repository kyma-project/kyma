---
title: Event name cleanup in Subscriptions
---

To conform to Cloud Event specifications, sometimes Eventing must modify the event names to filter out non-alphanumeric characters. This tutorial presents one example of event name cleanup.
You learn how Eventing behaves when you create a [Subscription](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) having alphanumeric characters in the event names. Read more about [Event name format and cleanup](../../05-technical-reference/evnt-01-event-names.md).

## Prerequisites

>**NOTE:** You need to enable Istio sidecar proxy injection, which is disabled on the Kyma clusters by default. To do this, follow the [Enable Istio Sidecar Injection](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md)

1. Follow the [prerequisites steps](../../02-get-started/04-trigger-workload-with-event.md#prerequisites) in the Getting Started guide.
2. Create a Function by following the [instructions](../../02-get-started/04-trigger-workload-with-event.md#create-a-function) in the Getting Started guide.
3. For this tutorial, instead of the default code sample, replace the Function source with the following code:

   <div tabs name="Deploy a Function" group="create-workload">
     <details open>
     <summary label="Kyma Dashboard">
     Kyma Dashboard
     </summary>

   ```js
   module.exports = {
     main: async function (event, context) {
       console.log("Received event: ", event.data, ", Event Type: ", event.extensions.request.headers['ce-type']);
       return;
     } 
   }
   ```
       
     </details>
     <details>
     <summary label="kubectl">
     kubectl
     </summary>
   
   ```bash
   cat <<EOF | kubectl apply -f -
     apiVersion: serverless.kyma-project.io/v1alpha1
     kind: Function
     metadata:
       labels:
         serverless.kyma-project.io/build-resources-preset: local-dev
         serverless.kyma-project.io/function-resources-preset: S
         serverless.kyma-project.io/replicas-preset: S
       name: lastorder
       namespace: default
     spec:
       deps: '{ "dependencies": {}}'
       maxReplicas: 1
       minReplicas: 1
       source: |
         module.exports = {
           main: async function (event, context) {
             console.log("Received event: ", event.data, ", Event Type: ", event.extensions.request.headers['ce-type']);
             return; 
           } 
         }
   EOF
   ```
   
     </details>
   </div>

## Create a Subscription with Event type consisting of alphanumeric characters

Next, create a [Subscription](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) custom resource and subscribe for events of the type: `order.payment-success.v1`. Note that `order.payment-success.v1` contains a non-alphanumeric character, the hyphen `-`.

<div tabs name="Create a Subscription" group="create-subscription">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, go to the view of your Function `lastorder`.
2. Go to **Configuration** > **Create Subscription+**.
3. Provide the following parameters:
   - **Subscription name**: `lastorder-payment-sub`
   - **Application name**: `myapp`
   - **Event name**: `order.payment-success`
   - **Event version**: `v1`

   - **Event type** is generated automatically. For this example, it's `sap.kyma.custom.myapp.order.payment-success.v1`.

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
   apiVersion: eventing.kyma-project.io/v1alpha1
   kind: Subscription
   metadata:
     name: lastorder-payment-sub
     namespace: default
   spec:
     sink: 'http://lastorder.default.svc.cluster.local'
     filter:
       filters:
         - eventSource:
             property: source
             type: exact
             value: ''
           eventType:
             property: type
             type: exact
             value: sap.kyma.custom.myapp.order.payment-success.v1
EOF
```

To check that the Subscription was created and is ready, run:
```bash
kubectl get subscriptions lastorder-payment-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the returned status says `true`.
  </details>
</div>

## Check the Subscription cleaned Event type

To check the Subscription cleaned Event type, run:
```bash
kubectl get subscriptions lastorder-payment-sub -o=jsonpath="{.status.cleanEventTypes}"
```

Note that the returned event type `["sap.kyma.custom.myapp.order.paymentsuccess.v1"]` does not contain the hyphen `-` in the `payment-success` part. That's because Kyma Eventing cleans out the non-alphanumeric characters from the event name and uses the cleaned event name in the underlying Eventing backend.

## Trigger the workload with an event

You created the `lastorder` Function, and subscribed to the `order.payment-success.v1` events by creating a Subscription CR. 
Next, you see that you can still publish events with the original Event name (i.e. `order.payment-success.v1`) even though it contains the non-alphanumeric character, and it will trigger the Function.

1. Port-forward the [Event Publisher Proxy](../../05-technical-reference/00-architecture/evnt-01-architecture.md) Service to localhost, using port `3000`. Run:
   ```bash
   kubectl -n kyma-system port-forward service/eventing-event-publisher-proxy 3000:80
   ```
2. Publish an event to trigger your Function. In another terminal window, run:

   <div tabs name="Publish an event" group="trigger-workload">
     <details open>
     <summary label="CloudEvents Conformance Tool">
     CloudEvents Conformance Tool
     </summary>
   
      ```bash
      cloudevents send http://localhost:3000/publish \
         --type sap.kyma.custom.myapp.order.payment-success.v1 \
         --id e4bcc616-c3a9-4840-9321-763aa23851fc \
         --source myapp \
         --datacontenttype application/json \
         --data "{\"orderCode\":\"3211213\", \"orderAmount\":\"1250\"}" \
         --yaml
      ```
   
     </details>
     <details>
     <summary label="curl">
     curl
     </summary>
   
      ```bash
      curl -v -X POST \
           -H "ce-specversion: 1.0" \
           -H "ce-type: sap.kyma.custom.myapp.order.payment-success.v1" \
           -H "ce-source: myapp" \
           -H "ce-eventtypeversion: v1" \
           -H "ce-id: e4bcc616-c3a9-4840-9321-763aa23851fc" \
           -H "content-type: application/json" \
           -d "{\"orderCode\":\"3211213\", \"orderAmount\":\"1250\"}" \
           http://localhost:3000/publish
      ```
     </details>
   </div>

## Verify the event delivery

To verify that the event was properly delivered, check the logs of the Function (see [Verify the event delivery](../../02-get-started/04-trigger-workload-with-event.md#verify-the-event-delivery)).

You see the received event in the logs:
```
Received event:  { orderCode: '3211213', orderAmount: '1250' } , Event Type:  sap.kyma.custom.myapp.order.paymentsuccess.v1
```
Note that the `Event Type` of the received event is not the same as defined in the Subscription.

## Conclusion

You see that Kyma Eventing modifies the event names to filter out non-alphanumeric character to conform to Cloud Event specifications. 

> **CAUTION:** This cleanup modification is abstract; you can still publish and subscribe to the original Event names. However, in some cases, it can lead to a naming collision as explained in [Event names](../../05-technical-reference/evnt-01-event-names.md).
