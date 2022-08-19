---
title: Trigger a workload with an event
---

We already know how to create and expose a workload ([Function](02-deploy-expose-function.md) and [microservice](03-deploy-expose-microservice.md)). 
Now it's time to actually use an event to trigger a workload.

## Prerequisites

1. Provision a [Kyma Cluster](01-quick-install.md).
2. (Optional) Deploy [Kyma Dashboard](../01-overview/main-areas/ui/ui-01-gui.md) on the Kyma cluster using the following command. Alternatively, you can also use `kubectl` CLI.
   ```bash
   kyma dashboard
   ```
3. (Optional) Install [CloudEvents Conformance Tool](https://github.com/cloudevents/conformance) for publishing events. Alternatively, you can also use `curl` to publish events.
   ```bash
   go install github.com/cloudevents/conformance/cmd/cloudevents@latest
   ```

## Create a Function

First, create a sample Function that prints out the received event to console:

>**NOTE:** You need to enable Istio sidecar proxy injection, which is disabled on the Kyma clusters by default. To do this, follow the [Enable Istio Sidecar Injection](../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md).

<div tabs name="Deploy a Function" group="trigger-workload">
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
  
  Run:

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

## Create a Subscription

Next, to subscribe to events, we need a [Subscription](../05-technical-reference/00-custom-resources/evnt-01-subscription.md) custom resource. We're going to subscribe to events of the type `order.received.v1`. 
All the published events of this type are then forwarded to an HTTP endpoint called `Sink`. You can define this endpoint in the Subscription's spec.

<div tabs name="Create a Subscription" group="trigger-workload">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, go to the view of your Function `lastorder`.
2. Go to **Configuration** > **Create Subscription+**.
3. Provide the following parameters:
   - **Subscription name**: `lastorder-sub`
   - **Application name**: `myapp`
   - **Event name**: `order.received`
   - **Event version**: `v1`

   - **Event type** is generated automatically. For this example, it's `sap.kyma.custom.myapp.order.received.v1`.

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
     filter:
       filters:
       - eventSource:
           property: source
           type: exact
           value: ""
         eventType:
           property: type
           type: exact
           value: sap.kyma.custom.myapp.order.received.v1
     sink: http://lastorder.default.svc.cluster.local
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

We created the `lastorder` Function and subscribed to the `order.received.v1` event by creating a Subscription CR. Now it's time to publish your event and trigger the Function. In this example, we'll port-forward the Kyma Eventing Service to localhost. 

1. Port-forward the Kyma Eventing Service to localhost. We will use port `3000`. Run: 
   ```bash
   kubectl -n kyma-system port-forward service/eventing-event-publisher-proxy 3000:80
   ```
2. Now publish an event to trigger your Function. In another terminal window, run: 

<div tabs name="Publish an event" group="trigger-workload">
  <details open>
  <summary label="CloudEvents Conformance Tool">
  CloudEvents Conformance Tool
  </summary>

   ```bash
   cloudevents send http://localhost:3000/publish \
      --type sap.kyma.custom.myapp.order.received.v1 \
      --id 759815c3-b142-48f2-bf18-c6502dc0998f \
      --source myapp \
      --datacontenttype application/json \
      --data "{\"orderCode\":\"3211213\"}" \
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
        -H "ce-id: 759815c3-b142-48f2-bf18-c6502dc0998f" \
        -H "content-type: application/json" \
        -d "{\"orderCode\":\"3211213\"}" \
        http://localhost:3000/publish
   ```
  </details>
</div>

## Verify the event delivery

To verify that the event was properly delivered, check the logs of the Function: 

<div tabs name="Verify the event delivery" group="trigger-workload">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, return to the view of your `lastorder` Function.
2. Go to **Code** and find the **Replicas of the Function** section.
3. Click on **View Logs**.
4. You see the received event in the logs:
   ```
   Received event: { orderCode: '3211213' }
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

You see the received event in the logs:
```
Received event: { orderCode: '3211213' }
```

  </details>
</div>

