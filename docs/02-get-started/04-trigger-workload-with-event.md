---
title: Trigger a workload with an event
---

We already know how to create and expose a workload ([Function](02-deploy-expose-function.md) and [microservice](03-deploy-expose-microservice.md)).
Now it's time to actually use an event to trigger a workload.

## Create a Function

First, create a sample Function that prints out the received event to console:

<!-- tabs:start -->

#### **Kyma Dashboard**

1. Go to **Namespaces** and select the `default` Namespace.
2. Go to **Workloads** > **Functions** and click **Create Function +**.
3. Name the Function `lastorder`.
4. From the **Language** dropdown, choose `JavaScript`.
5. From the **Runtime** dropdown, choose one of the available `nodejs`.
6. In the **Source** section, replace its source with the following code:
    ```js
    module.exports = {
      main: async function (event, context) {
        console.log("Received event:", event.data);
        return;
      }
    }
    ```
7. Click **Create**.
8. Wait a few seconds for the Function to have the status `RUNNING`.

#### **kubectl**

Run:

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
            console.log("Received event:", event.data);
            return;
          }
        }
EOF
```

The Function was created successfully if the command returns this message:

```bash
function.serverless.kyma-project.io/lastorder created
```

Wait a few seconds for the Function to have status `RUNNING`. To check the Function status, run:

```bash
kubectl get functions -n default lastorder
```

> **NOTE:** You might need to wait a few seconds for the Function to be ready.

<!-- tabs:end -->

## Create a Subscription

Next, to subscribe to events, we need a Subscription custom resource. We're going to subscribe to events of the type `order.received.v1`.
All the published events of this type are then forwarded to an HTTP endpoint called `Sink`. You can define this endpoint in the Subscription's spec.

<!-- tabs:start -->

#### **Kyma Dashboard**

1. Go to **Namespaces** and select the `default` Namespace.
2. Go to **Configuration** > **Subscriptions** and click **Create Subscription+**.
3. Provide the following parameters:
   - **Subscription name**: `lastorder-sub`
   - **Types**: `order.received.v1`
   - **Service**: `lastorder` (The sink field will be populated automatically.)
   - **Type matching**: `standard`
   - **Source**: `myapp`

4. Click **Create**.
5. Wait a few seconds for the Subscription to have status `READY`.

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
     source: myapp
     types:
       - order.received.v1
     sink: http://lastorder.default.svc.cluster.local
EOF
```

To check that the Subscription was created and is ready, run:
```bash
kubectl get subscriptions lastorder-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the returned status says `true`.

<!-- tabs:end -->

## Trigger the workload with an event

We created the `lastorder` Function and subscribed to the `order.received.v1` event by creating a Subscription CR. Now it's time to publish your event and trigger the Function. In this example, we'll port-forward the Kyma Eventing Service to localhost.

1. Port-forward the Kyma Eventing Service to localhost. We will use port `3000`. In your terminal, run:
   ```bash
   kubectl -n kyma-system port-forward service/eventing-event-publisher-proxy 3000:80
   ```
2. Now publish an event to trigger your Function. In another terminal window, run:

<!-- tabs:start -->

#### **curl**

   ```bash
   curl -v -X POST \
        -H "ce-specversion: 1.0" \
        -H "ce-type: order.received.v1" \
        -H "ce-source: myapp" \
        -H "ce-eventtypeversion: v1" \
        -H "ce-id: 759815c3-b142-48f2-bf18-c6502dc0998f" \
        -H "content-type: application/json" \
        -d "{\"orderCode\":\"3211213\"}" \
        http://localhost:3000/publish
   ```

 #### **CloudEvents Conformance Tool**

   ```bash
   cloudevents send http://localhost:3000/publish \
      --type order.received.v1 \
      --id 759815c3-b142-48f2-bf18-c6502dc0998f \
      --source myapp \
      --datacontenttype application/json \
      --data "{\"orderCode\":\"3211213\"}" \
      --yaml
   ```

<!-- tabs:end -->

## Verify the event delivery

To verify that the event was properly delivered, check the logs of the Function:

<!-- tabs:start -->

#### **Kyma Dashboard**

1. In Kyma Dashboard, return to the view of your `lastorder` Function.
2. In the **Code** view, find the **Replicas of the Function** section.
3. Click the name of your replica.
4. Locate the **Containers** section and click on **View Logs**.
5. You see the received event in the logs:
   ```
   Received event: { orderCode: '3211213' }
   ```

#### **kubectl**

Run:

```bash
kubectl logs \
  -n default \
  -l serverless.kyma-project.io/function-name=lastorder,serverless.kyma-project.io/resource=deployment \
  -c function
```

You see the received event in the logs:
```
Received event: { orderCode: '3211213' }
```

<!-- tabs:end -->

That's it!

Go ahead and dive a little deeper into the Kyma documentation for [tutorials](../03-tutorials), [operation guides](../04-operation-guides), and [technical references](../05-technical-reference), as well as information on the [main areas in Kyma](../01-overview/). Happy Kyma-ing!
