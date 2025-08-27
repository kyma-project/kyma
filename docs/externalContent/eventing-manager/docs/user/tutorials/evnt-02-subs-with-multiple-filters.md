# Create Subscription Subscribing to Multiple Event Types

The [Subscription](../resources/evnt-cr-subscription.md) CustomResourceDefinition (CRD) is used to subscribe to events. In this tutorial, you learn how to subscribe to one or more event types using the Kyma Subscription.

## Prerequisites

> [!NOTE]
> Read about the [Purpose and Benefits of Istio Sidecar Proxies](https://kyma-project.io/#/istio/user/00-00-istio-sidecar-proxies?id=purpose-and-benefits-of-istio-sidecar-proxies). Then, check how to [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection). For more details, see [Default Istio Configuration](https://kyma-project.io/#/istio/user/00-15-overview-istio-setup) in Kyma.

1. Follow the [Prerequisites steps](evnt-01-prerequisites.md) for the Eventing tutorials.
2. [Create and Modify an Inline Function](https://kyma-project.io/#/serverless-manager/user/tutorials/01-10-create-inline-function).

## Create a Subscription With Multiple Event Types

To subscribe to multiple events, you need a [Subscription](../resources/evnt-cr-subscription.md) custom resource (CR). In the following example, you learn how to subscribe to events of two types: `order.received.v1` and `order.changed.v1`.

<Tabs>
<Tab name="Kyma Dashboard">

1. Go to **Namespaces** and select the default namespace.
2. Go to **Configuration** > **Subscriptions** and click **Create Subscription+**.
3. Provide the following parameters:
   - **Subscription name**: `lastorder-sub`
   - **Types**: `order.received.v1` and `order.changed.v1`
   - **Service**: `lastorder` (The sink field will be populated automatically.)
   - **Type matching:**: `standard`
   - **Source**: `myapp`

   > **NOTE:** You can add more types to your subscription if you want to subscribe to more event types.

4. Click **Create**.
5. Wait a few seconds for the Subscription to have status `READY`.
</Tab>
<Tab name="kubectl">

Run:

```bash
cat <<EOF | kubectl apply -f -
    apiVersion: eventing.kyma-project.io/v1alpha2
    kind: Subscription
    metadata:
      name: lastorder-sub
      namespace: default
    spec:
      sink: 'http://lastorder.default.svc.cluster.local'
      source: myapp
      types:
       - order.received.v1
       - order.changed.v1
EOF
```

To check that the Subscription was created and is ready, run:

```bash
kubectl get subscriptions lastorder-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the returned status says `true`.
</Tab>
</Tabs>

## Trigger the Workload with an Event

You created the `lastorder` Function, and subscribed to the `order.received.v1` and `order.changed.v1` events by creating a Subscription CR. Now it's time to publish the events and trigger the Function.
In the following example, you port-forward the [Eventing Publisher Proxy](../evnt-architecture.md) Service to localhost.

1. Port-forward the [Eventing Publisher Proxy](../evnt-architecture.md) Service to localhost, using port `3000`. Run:

   ```bash
   kubectl -n kyma-system port-forward service/eventing-publisher-proxy 3000:80
   ```

2. Publish an event of type `order.received.v1` to trigger your Function. In another terminal window, run:

<Tabs>
<Tab name="CloudEvents Conformance Tool">

```bash
cloudevents send http://localhost:3000/publish \
   --type order.received.v1 \
   --id cc99dcdd-6f6d-43d6-afef-d024eb276584 \
   --source myapp \
   --datacontenttype application/json \
   --data "{\"orderCode\":\"3211213\", \"orderStatus\":\"received\"}" \
   --yaml
```
</Tab>
<Tab name="curl">

```bash
curl -v -X POST \
     -H "ce-specversion: 1.0" \
     -H "ce-type: order.received.v1" \
     -H "ce-source: myapp" \
     -H "ce-eventtypeversion: v1" \
     -H "ce-id: cc99dcdd-6f6d-43d6-afef-d024eb276584" \
     -H "content-type: application/json" \
     -d "{\"orderCode\":\"3211213\", \"orderStatus\":\"received\"}" \
     http://localhost:3000/publish
```
</Tab>
</Tabs>

3. Now, publish an event of type `order.changed.v1` to trigger your Function.

<Tabs>
<Tab name="Kyma Dashboard">

```bash
cloudevents send http://localhost:3000/publish \
   --type order.changed.v1 \
   --id 94064655-7e9e-4795-97a3-81bfd497aac6 \
   --source myapp \
   --datacontenttype application/json \
   --data "{\"orderCode\":\"3211213\", \"orderStatus\":\"changed\"}" \
   --yaml
```
</Tab>
<Tab name="CloudEvents Conformance Tool">

```bash
curl -v -X POST \
     -H "ce-specversion: 1.0" \
     -H "ce-type: order.changed.v1" \
     -H "ce-source: myapp" \
     -H "ce-eventtypeversion: v1" \
     -H "ce-id: 94064655-7e9e-4795-97a3-81bfd497aac6" \
     -H "content-type: application/json" \
     -d "{\"orderCode\":\"3211213\", \"orderStatus\":\"changed\"}" \ 
     http://localhost:3000/publish
```
</Tab>
</Tabs>

## Verify the Event Delivery

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
Received event: { orderCode: '3211213', orderStatus: 'received' }
Received event: { orderCode: '3211213', orderStatus: 'changed' }
```
