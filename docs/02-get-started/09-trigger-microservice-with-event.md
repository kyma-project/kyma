---
title: Trigger the microservice with an event
type: Getting Started
---

This tutorial shows how to trigger the deployed `orders-service` microservice with the `order.deliverysent.v1` event from Commerce mock previously connected to Kyma.

## Reference

This guide demonstrates how [Eventing](/components/eventing/) works in Kyma. It allows you to receive business events from external solutions and trigger business flows with Functions or microservices.

## Steps

### Create the Subscription

Follow these steps:

<div tabs name="steps" group="subscribe-microservice">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create a Subscription custom resource (CR) to subscribe it to the `order.deliverysent.v1` event type from Commerce mock:

   ```bash
   cat <<EOF | kubectl apply -f  -
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
              value: sap.kyma.custom.commercemock.order.deliverysent.v1
        protocol: ""
        protocolsettings: {}
        sink: http://orders-function.orders-service.svc.cluster.local
   EOF
   ```

The event type is composed of the following components:
- Prefix: `sap.kyma.custom`
- Application: `commerce`
- Event: `order.deliverysent`
- Version: `v1`

2. Check that the Subscription CR was created and is ready. This is indicated by its status equal to `true`:

   ```bash
   kubectl get subscriptions.eventing.kyma-project.io orders-sub -n orders-service -o=jsonpath="{.status.ready}"
   ```

   </details>
<details>
<summary label="console-ui">
Console UI
</summary>

1. From the drop-down list in the top navigation panel, select the `orders-service` Namespace.

2. Go to **Discovery and Network** > **Services** in the left navigation panel and select `orders-service`.

3. Once in the service's details view, switch to the **Configuration** tab and select **Create Event Subscription** in the **Event Subscriptions** section.

4. Find the `order.deliverysent` event type with the `v1` version from the `commerce-mock` application. Mark it on the list and select **Add**.

A message confirming that the Subscription was created will appear in the **Event Subscriptions** section in the service's details view.

  </details>
</div>


### Test the event delivery

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

   [{"orderCode":"123456789","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```
