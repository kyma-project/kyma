---
title: Trigger the Function with an event
type: Getting Started
---

As the final step, you will trigger the Function with the `order.deliverysent` event type from Commerce mock, send a sample event from the mock application, and test if the event reached the Function.

## Steps

### Create the Trigger

Follows these steps:

<div tabs name="steps" group="trigger-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create a Trigger CR for `orders-function` to subscribe the Function to the `order.deliverysent` event type from Commerce mock:

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
- **spec.filter.attributes.eventtypeversion** points to the specific event version. In this example, it is `v1`.
- **spec.filter.attributes.source** is taken from the name of the Application CR and specifies the source of events. In this example, it is `commerce-mock`.
- **spec.filter.attributes.type** points to the given event type to which you want to subscribe the Function. In this example, it is `order.deliverysent`.

2. Check if the Trigger CR was created and is ready. Its status should be `True`:

  ```bash
  kubectl get trigger orders-function -n orders-service -o=jsonpath="{.status.conditions[2].status}"
  ```

    </details>
    <details>
    <summary label="ui">
    UI
    </summary>

1. Select the `orders-service` Namespace from the drop-down list in the top navigation panel.

2. Go to **Development** > **Functions** in the left navigation panel and navigate to `orders-function`.

3. Once in the Function's details view, switch to the **Configuration** tab and select **Add Event Trigger** in the **Event Triggers** section.

4. Once the pop-up box opens, find the `order.deliverysent` event with the `v1` version from the `commerce-mock` application. Mark it on the list and select **Add**.

The message will appear in the UI confirming that the event trigger was created, and you will see it in the **Event Triggers** section in the Function's details view.

    </details>
</div>

### Test the Trigger

To send events from Commerce mock to `orders-function`, follow these steps:

1. Access Commerce mock at `https://commerce-orders-service.{CLUSTER_DOMAIN}.` or use the link under **Host** in the **Configuration** > **API Rules** view in the `order-service` Namespace.

2. Switch to the **Remote APIs** tab, find **SAP Commerce Cloud - Events** and select it.

3. Select the `order.deliverysent.v1` event type in **Event Topics** drop-down list. In the details of the printed event, change **orderCode** to `987654321` and select **Send Event**.

   The message appears on the UI confirming that the event was sent.

4. Call the Function to verify if the event details were saved:

   ```bash
   curl -ik "https://$FUNCTION_DOMAIN"
   ```

   > **NOTE**: To get the domain of the Function, run:
   >
   > ```bash
   > export FUNCTION_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-function.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
   > ```

   You should see a similar response proving that the `987654321` event was delivered as expected:

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

   [{"orderCode":"987654321","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"},
   {"orderCode":"762727234","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"123456789","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```
