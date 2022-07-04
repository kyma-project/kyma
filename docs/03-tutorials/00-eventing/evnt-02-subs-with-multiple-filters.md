---
title: Create subscription subscribing to multiple event types
---

The [Subscription](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) custom resource definition (CRD) is used to subscribe to events. In this tutorial, we will show how to subscribe to one or more event types using the Kyma Subscription.

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
4. In the inline editor for the Function, modify its source replacing it with this code:
    ```js
    module.exports = {
      main: async function (event, context) {
        console.log("Received event:", event.data);
        return;
      } 
    }
    ```
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
          console.log("Received event:", event.data);
          return; 
        } 
      }
EOF
```

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

## Create a Subscription with Multiple Filters

Next, to subscribe to multiple events, we need a [Subscription](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) custom resource. We're going to subscribe to events of two types: `order.received.v1` and `order.changed.v1`.
All the published events of these types are then forwarded to an HTTP endpoint called `Sink`. You can define this endpoint in the Subscription's spec.

<div tabs name="Create a Subscription" group="create-subscription">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, go to the view of your Function `lastorder`.
2. Go to **Configuration** > **Create Subscription+**.
3. Switch to the **Advanced** tab, and name the Subscription `lastorder-sub`.
4. Add a second Filter using **Filters** > **Add Filter +**.
5. Provide the `Event type` under `Filter 1` as `sap.kyma.custom.myapp.order.received.v1`. Leave the `Event source` as empty.
6. Provide the `Event type` under `Filter 2` as `sap.kyma.custom.myapp.order.changed.v1`. Leave the `Event source` as empty.

   > **NOTE:** You can add more filters to your subscription if you want to subscribe to more event types.

7. Click **Create**.
8. Wait a few seconds for the Subscription to have status `READY`.

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
          - eventSource:
              property: source
              type: exact
              value: ''
            eventType:
              property: type
              type: exact
              value: sap.kyma.custom.myapp.order.changed.v1
EOF
```

To check that the Subscription was created and is ready, run:
```bash
kubectl get subscriptions lastorder-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the returned status says `true`.

  </details>
</div>

## Trigger the workload with an event

We created the `lastorder` Function, and subscribed to the `order.received.v1` and `order.changed.v1` events by creating a Subscription CR. Now it's time to publish the events and trigger the Function.
In this example, we'll port-forward the Kyma Eventing Service to localhost.

1. Port-forward the Kyma Eventing Service to localhost. We will use port `3000`. Run:
   ```bash
   kubectl -n kyma-system port-forward service/eventing-event-publisher-proxy 3000:80
   ```
2. Publish an event of type `order.received.v1` to trigger your function. In another Terminal window run:

    <div tabs name="Publish an event" group="trigger-workload">
      <details open>
      <summary label="CloudEvents Conformance Tool">
      CloudEvents Conformance Tool
      </summary>
    
       ```bash
       cloudevents send http://localhost:3000/publish \
          --type sap.kyma.custom.myapp.order.received.v1 \
          --id cc99dcdd-6f6d-43d6-afef-d024eb276584 \
          --source myapp \
          --datacontenttype application/json \
          --data "{\"orderCode\":\"3211213\", \"orderStatus\":\"received\"}" \
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
            -H "ce-type: sap.kyma.custom.myapp.order.received.v1" \
            -H "ce-source: /default/io.kyma-project/custom" \
            -H "ce-eventtypeversion: v1" \
            -H "ce-id: cc99dcdd-6f6d-43d6-afef-d024eb276584" \
            -H "content-type: application/json" \
            -d "{\"orderCode\":\"3211213\", \"orderStatus\":\"received\"}" \
            http://localhost:3000/publish
       ```
      </details>
    </div>

3. Now publish an event of type `order.changed.v1` to trigger your function.

    <div tabs name="Publish an event" group="trigger-workload2">
      <details open>
      <summary label="CloudEvents Conformance Tool">
      CloudEvents Conformance Tool
      </summary>
    
       ```bash
       cloudevents send http://localhost:3000/publish \
          --type sap.kyma.custom.myapp.order.changed.v1 \
          --id 94064655-7e9e-4795-97a3-81bfd497aac6 \
          --source myapp \
          --datacontenttype application/json \
          --data "{\"orderCode\":\"3211213\", \"orderStatus\":\"changed\"}" \
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
            -H "ce-type: sap.kyma.custom.myapp.order.changed.v1" \
            -H "ce-source: /default/io.kyma-project/custom" \
            -H "ce-eventtypeversion: v1" \
            -H "ce-id: 94064655-7e9e-4795-97a3-81bfd497aac6" \
            -H "content-type: application/json" \
            -d "{\"orderCode\":\"3211213\", \"orderStatus\":\"changed\"}" \
            http://localhost:3000/publish
       ```
      </details>
    </div>

## Verify the event delivery

To verify that the events were properly delivered, check the logs of the Function:

<div tabs name="Verify the event delivery" group="verify-event">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, return to the view of your `lastorder` Function.
2. Go to **Code** and find the **Replicas of the Function** section.
3. Click on **View Logs**.
4. You will see the received event in the logs:
   ```
   Received event: { orderCode: '3211213', orderStatus: 'received' }
   Received event: { orderCode: '3211213', orderStatus: 'changed' }
   ```

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

You will see the received event in the logs:
```
Received event: { orderCode: '3211213', orderStatus: 'received' }
Received event: { orderCode: '3211213', orderStatus: 'changed' }
```

  </details>
</div>
