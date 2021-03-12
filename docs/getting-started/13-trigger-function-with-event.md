---
title: Trigger the Function with an event
type: Getting Started
---

As the final step, you will trigger the Function with the `order.created.v1` event type from Commerce mock, send a sample event from the mock application, and test if the event reached the Function.

## Steps

### Create the Kyma Subscription

Follows these steps:

<div tabs name="steps" group="function-subscription">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create a Kyma Subscription custom resource (CR) to subscribe the Function to the `order.created.v1` event type from Commerce mock:

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: Subscription
metadata:
  name: orders-sub
  namespace: orders-service
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
        value: sap.kyma.custom.commerce.order.created.v1
  protocol: ""
  protocolsettings: {}
  sink: http://orders-function.orders-service.svc.cluster.local
```

The event type is composed of the following components:
- Prefix: `sap.kyma.custom`
- Application: `commerce`
- Event: `order.created`
- Version: `v1`

2. Check that the Subscription CR was created and is ready. This is indicated by its status equal to `true`:

  ```bash
  kubectl get subscriptions.eventing.kyma-project.io orders-function -n orders-service -o=jsonpath="{.status.ready}"
  ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. From the drop-down list in the top navigation panel, select the `orders-service` Namespace.

2. Go to **Workloads** > **Functions** in the left navigation panel and navigate to `orders-function`.

3. Once in the Function's details view, switch to the **Configuration** tab and select **Create Event Subscriptions** in the **Event Subscriptions** section.

4. Once the pop-up box opens, find the `order.created.v1` event with the `v1` version from the `commerce-mock` application. Check it on the list and select **Add**.

A message confirming that the Subscription was created will appear in the **Event Subscriptions** section in the Function's details view.

    </details>
</div>

### Test the Event delivery

To send events from Commerce mock to `orders-function`, follow these steps:

1. Access Commerce mock at `https://commerce-orders-service.{CLUSTER_DOMAIN}.` or use the link under **Host** in the **Discovery and Network** > **API Rules** view in the `order-service` Namespace.

2. Switch to the **Remote APIs** tab, find **SAP Commerce Cloud - Events**, and select it.

3. From the **Event Topics** drop-down list, select the `order.created.v1` event type. In the details of the printed event, change **orderCode** to `987654321` and select **Send Event**.

   A message confirming that the event was sent will appear in the UI.

4. Call the Function to verify that the event details were saved:

   ```bash
   curl -ik "https://$FUNCTION_DOMAIN"
   ```

   > **NOTE**: To get the domain of the Function, run:
   >
   > ```bash
   > export FUNCTION_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-function.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
   > ```

   The system returns the response proving that the `987654321` event was delivered as expected:

   ```bash
   HTTP/2 200
   access-control-allow-origin: *
   content-length: 652
   content-type: application/json; charset=utf-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   etag: W/"28c-MLZh1MyovyUrCPwMzfRWfVQwhlU"
   server: istio-envoy
   x-envoy-upstream-service-time: 991
   x-powered-by: Express

   [{"orderCode":"987654321","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```
