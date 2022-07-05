---
title: Changing Events Max-In-Flight in Subscriptions
---

In this tutorial, we will show how to set idle "in-flight messages" limit in Kyma Subscriptions.
The "in-flight messages" config defines the number of events which will be forwarded in parallel to the sink by Eventing Services without waiting for a response. 

## Prerequisites

1. Provision a [Kyma Cluster](../../02-get-started/01-quick-install.md).
2. (Optional) Deploy [Kyma Dashboard](../../01-overview/main-areas/ui/ui-01-gui.md) on the Kyma cluster using the following command. Alternatively, you can also use `kubectl` CLI.
   ```bash
   kyma dashboard
   ```
3. (Optional) Install [CloudEvents Conformance Tool](https://github.com/cloudevents/conformance) for publishing events. Alternatively, you can also use `curl` to publish events.
   ```bash
   go install github.com/cloudevents/conformance/cmd/cloudevents@latest
   ```

## Create a Workload

First, create a sample Function that prints out the received event to console:

<div tabs name="Deploy a Function" group="create-workload">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. Go to **Namespaces** and select the default Namespace.
2. Go to **Workloads** > **Functions** and click **Create Function +**.
3. Name the Function `lastorder` and click **Create**.
4. In the inline editor for the Function, replace its source with the following code:
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
   The Function will wait for 5 seconds before returning the response in order to simulate prolonged event processing.
5. Save your changes.
6. Wait a few seconds for the Function to have status `RUNNING`.

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
          console.log("Processing event:", event.data);
          // sleep/wait for 5 seconds
          await new Promise(r => setTimeout(r, 5 * 1000));
          console.log("Completely processed event:", event.data);
          return;
        } 
      }
EOF
```

The Function will wait for 5 seconds before returning the response in order to simulate prolonged event processing.

<br/>
If the resources were created successfully, the command returns this message:

```bash
function.serverless.kyma-project.io/lastorder created
```

To check the Function status, run:

```bash
kubectl get functions -n default lastorder
```

> **NOTE:** You might need to wait a few seconds for the Function to be ready.

  </details>
</div>

## Create a Subscription with Max-In-Flight config

Next, we will create a [Subscription](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) custom resource. We're going to subscribe for events of the type: `order.received.v1`. We will set the `maxInFlightMessages` to `5`, so that the Eventing Services forwards maximum 5 events in parallel to the sink without waiting for a response.

<div tabs name="Create a Subscription" group="create-subscription">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, go to the view of your Function `lastorder`.
2. Go to **Configuration** > **Create Subscription+**.
3. Switch to the **Advanced** tab, and replace the yaml with the following:
   ```bash
   apiVersion: eventing.kyma-project.io/v1alpha1
   kind: Subscription
   metadata:
     name: lastorder-sub
     namespace: default
   spec:
     config:
       maxInFlightMessages: 5
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
             value: sap.kyma.custom.myapp.order.received.v1
   ```

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
     name: lastorder-sub
     namespace: default
   spec:
     config:
       maxInFlightMessages: 5
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
             value: sap.kyma.custom.myapp.order.received.v1
EOF
```

To check that the Subscription was created and is ready, run:
```bash
kubectl get subscriptions lastorder-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the returned status says `true`.
  </details>
</div>

## Trigger the workload with multiple events

We created the `lastorder` Function, and subscribed to the `order.received.v1` events by creating a Subscription CR.
We will publish 15 events at once and see how the Eventing Service trigger the workload.
In this example, we'll port-forward the [Event Publisher Proxy](../../05-technical-reference/00-architecture/evnt-01-architecture.md) Service to localhost.

1. Port-forward the Event Publisher Proxy Service to localhost. We will use port `3000`. Run:
   ```bash
   kubectl -n kyma-system port-forward service/eventing-event-publisher-proxy 3000:80
   ```
2. Now publish 15 events to the Eventing Service. In another Terminal window run:

   <div tabs name="Publish an event" group="trigger-workload">
     <details open>
     <summary label="CloudEvents Conformance Tool">
     CloudEvents Conformance Tool
     </summary>
   
     ```bash
     for i in {1..15}
     do
       cloudevents send http://localhost:3000/publish \
         --type sap.kyma.custom.myapp.order.received.v1 \
         --id e4bcc616-c3a9-4840-9321-763aa23851f${i} \
         --source myapp \
         --datacontenttype application/json \
         --data "{\"orderCode\":\"$i\"}" \
         --yaml
     done
     ```
   
     </details>
     <details>
     <summary label="curl">
     curl
     </summary>
   
     ```bash
     for i in {1..15}
     do
       curl -v -X POST \
         -H "ce-specversion: 1.0" \
         -H "ce-type: sap.kyma.custom.myapp.order.received.v1" \
         -H "ce-source: myapp" \
         -H "ce-eventtypeversion: v1" \
         -H "ce-id: e4bcc616-c3a9-4840-9321-763aa23851f${i}" \
         -H "content-type: application/json" \
         -d "{\"orderCode\":\"$i\"}" \
         http://localhost:3000/publish
     done
     ```
     </details>
   </div>

## Verify the event delivery

To verify that the events ware properly delivered, check the logs of the Function:

<div tabs name="Verify the event delivery" group="trigger-workload">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, return to the view of your `lastorder` Function.
2. Go to **Code** and find the **Replicas of the Function** section.
3. Click on **View Logs**.

</details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>
Run: 

```bash
kubectl logs -f -n default \
  $(kubectl get pod \
    --field-selector=status.phase==Running \
    -l serverless.kyma-project.io/function-name=lastorder \
    -o jsonpath="{.items[0].metadata.name}")
```
  </details>
</div>

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

You can see that only 5 events at maximum were delivered to the Function in parallel and as soon as the Function completes the processing of the event and returns the response, the Eventing Service delivers the next in-line event to the Function. 
