---
title: Trigger the microservice with an event
type: Getting Started
---

This tutorial shows how to trigger the deployed `orders-service` microservice with the `order.deliverysent.v1` event from Commerce mock previously connected to Kyma.

## Reference

This guide demonstrates how [Event Mesh](/components/event-mesh/) works in Kyma. It allows you to receive business events from external solutions and trigger business flows with Functions or microservices.

## Steps

### Create the event trigger

Follow these steps:

<div tabs name="steps" group="trigger-microservice">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Run the `kubectl get brokers -n {NAMESPACE}` command to check if there already is the Knative's `default` Broker running in the Namespace where your Function is running. If not, you must manually inject the Broker into the Namespace to enable Trigger creation and event flow. To do that, run this command:

  ```bash
  kubectl label namespace {NAMESPACE} knative-eventing-injection=enabled
  ```

2. Create a [Trigger CR](https://knative.dev/docs/eventing/triggers/) for the `orders-service` microservice to subscribe it to the `order.deliverysent.v1` event type from Commerce mock:

  ```yaml
  cat <<EOF | kubectl apply -f  -
    apiVersion: eventing.knative.dev/v1alpha1
    kind: Trigger
    metadata:
      name: orders-service
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
          name: orders-service
          namespace: orders-service
  EOF
  ```

- **spec.filter.attributes.eventtypeversion** points to the specific event version type. In this example, it is `v1`.
- **spec.filter.attributes.source** is the name of the Application CR which is the source of the events. In this example, it is `commerce-mock`.
- **spec.filter.attributes.type** points to the event type to which you want to subscribe the microservice. In this example, it is `order.deliverysent`.

3. Check that the Trigger CR was created and is ready. This is indicated by its status equal to `True`:

   ```bash
   kubectl get trigger orders-service -n orders-service -o=jsonpath="{.status.conditions[2].status}"
   ```

   </details>
<details>
<summary label="console-ui">
Console UI
</summary>

1. From the drop-down list in the top navigation panel, select the `orders-service` Namespace.

2. Go to **Discovery and Network** > **Services** in the left navigation panel and select `orders-service`.

3. Once in the service's details view, select **Add Event Trigger** in the **Event Triggers** section.

4. Find the `order.deliverysent` event type with the `v1` version from the `commerce-mock` application. Mark it on the list and select **Add**.

   A message confirming that the event trigger was created will appear in the **Event Triggers** section of the service's details view.

  </details>
</div>


### Test the trigger

To send events from Commerce mock to the `orders-service` microservice, follow these steps:  

1. Access Commerce mock at `https://commerce-orders-service.{CLUSTER_DOMAIN}.` or use the link under **Host** in the **Discovery and Network** > **API Rules** view in the `order-service` Namespace.

2. Switch to the **Remote APIs** tab, find **SAP Commerce Cloud - Events**, and select it.

3. From the **Event Topics** drop-down list, select the `order.deliverysent.v1` event type.

4. In the details of the printed event, change **orderCode** to `123456789` and select **Send Event**.

   A message will appear in the UI confirming that the event was sent.

5. Call the microservice to verify that the event details were saved:

   ```bash
   curl -ik "https://$SERVICE_DOMAIN/orders"
   ```

   > **NOTE**: To get the domain of the microservice, run:
   >
   > ```bash
   > export SERVICE_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-service.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
   > ```

   The system returns a response proving that the microservice received the `123456789` order details:

   ```bash
   content-length: 2
   content-type: application/json;charset=UTF-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   server: istio-envoy
   vary: Origin
   x-envoy-upstream-service-time: 37

   [{"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"123456789","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```
