---
title: Trigger a Function with an event
type: Getting Started
---

This tutorial shows how to trigger a Function with an event from Commerce mock connected to Kyma.

> **NOTE:** To learn more about events flow in Kyma, read the [eventing](/components/event-mesh) documentation.

## Steps

### Create the Trigger

Follows these steps:

<div tabs name="steps" group="trigger-function">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Create a Trigger CR for the `orders-function` Function to subscribe the Function to the `order.deliverysent.v1` event Commerce mock:

```yaml
cat <<EOF | kubectl apply -f  -
apiVersion: eventing.knative.dev/v1alpha1
kind: Trigger
metadata:
  name: orders-function
  namespace: orders-service
spec:
  broker: default
  filter:
    attributes:
      eventtypeversion: v1
      source: commerce-mock
      type: order.deliverysent
  subscriber:
    ref:
      apiVersion: v1
      kind: Service
      name: orders-function
      namespace: orders-service
EOF
```

where:
- **spec.filter.attributes.eventtypeversion** points to the specific event version. In our case, it is `v1`.
- **spec.filter.attributes.source** is taken from the name of the Application CR and specifies the source of events. In our example, it is `commerce-mock`.
- **spec.filter.attributes.type** points to the given event type to which you want to subscribe the Function. In our case, it is `order.deliverysent`.

2. Check if the Trigger CR was created and is ready. The status of the CR should state `True`:

   ```bash
   kubectl get trigger orders-function -n orders-service -o=jsonpath="{.status.conditions[2].status}"
   ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. Navigate to the `orders-service` Namespace view in the Console UI from the drop-down list in the top navigation panel.

2. Go to the **Functions** view under the **Development** section in the left navigation panel and navigate to `orders-function`.

3. Once in the Function's details view, switch to the **Configuration** tab and select **Add Event Trigger** in the **Event Triggers** section.

4. Once the pop-up box opens, find the `order.deliverysent` event with the `v1` version from the `commerce-mock` application. Mark it on the list and select **Add**.

The message appears on the UI confirming that the event trigger was created, and you will see it in the **Event Triggers** section in the Function's details view.

    </details>
</div>

_TO_DO_BELOW_
## Test the Trigger

To send events from mock to Orders Service application, follow these steps:  

1. Access the SAP Commerce Cloud Mock mock at `https://commerce-orders-service.{CLUSTER_DOMAIN}.` or go to **API Rules** view (under **Configuration** section) in `orders-service` Namespace and select the mock, you will the direct link to the mock application under **Host** column.

2. Switch to **Remote APIs** tab, find **SAP Commerce Cloud - Events** and click it.

3. In opened view search in dropdown list `order.deliverysent.v1` event. In pasted event change `orderCode` to `987654321` and select **Send Event**.

   The message appears on the UI confirming that the event was successfully sent.

4. For the last time call the Function to check the storage:

   ```bash
   curl -ik "https://$FUNCTION_DOMAIN"
   ```

   > **NOTE**: To get the Function domain, run:
   >
   > ```bash
   > export FUNCTION_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-function.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
   > ```

   You should see a response similar to the following:

   ```bash
   content-length: 2
   content-type: application/json;charset=UTF-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   server: istio-envoy
   vary: Origin
   x-envoy-upstream-service-time: 37

   [{"orderCode":"762727234","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"123456789","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"987654321","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```

   The event from mock application was saved in Redis instance :)
