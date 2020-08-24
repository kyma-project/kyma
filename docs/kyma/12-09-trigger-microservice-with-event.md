---
title: Trigger a microservice with an event
type: Getting Started
---

This tutorial shows how to trigger the deployed `orders-service` microservice with the `order.deliverysent.v1` event from Commerce mock connected to Kyma.

## Related Kyma components

This guide demonstrates how [Event Mesh](/components/event-mesh/) works in Kyma. It allows you receives business events from external solutions and is able to trigger business flows using Functions or microservices.

## Steps

### Create the event trigger

<div tabs name="steps" group="trigger-microservice">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Create the Trigger CR for the `orders-service` microservice to subscribe it to the `order.deliverysent.v1` event from Commerce mock:

```bash
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

- **spec.filter.attributes.eventtypeversion** points to the specific event version, on our case it is `v1`.
- **spec.filter.attributes.source** is taken from the name of the Application CR and specifies the source of events. In our example, it is `commerce-mock`.
- **spec.filter.attributes.type** points to the given event type to which you want to subscribe the microservice. In our case, it is `order.deliverysent`.

2. Check if the Trigger CR was created and is ready. The status of the CR should state `True`:

   ```bash
   kubectl get trigger orders-service -n orders-service -o=jsonpath="{.status.conditions[2].status}"
   ```

   </details>
<details>
<summary label="console-ui">
Console UI
</summary>

1. Navigate to the `orders-service` Namespace view in the Console UI from the drop-down list in the top navigation panel.

2. Go to the **Services** view under the **Operation** section in the left navigation panel and navigate to `orders-service`.

3. Once in the service's details view, select **Add Event Trigger** in the **Event Triggers** section.

4. Find the `order.deliverysent` event with the `v1` version from the `commerce-mock` application. Mark it on the list and select **Add**.

   The message appears on the UI confirming that the event trigger was created, and you will see it in the **Event Triggers** section of service's details view.

  </details>
</div>


### Test the trigger

To send events from Commerce mock to the `orders-service` microservice, follow these steps:  

1. Access Commerce mock at `https://commerce-orders-service.{CLUSTER_DOMAIN}.` or go to **API Rules** view (under **Configuration** section) in the `orders-service` Namespace and select the mock application. You can also follow the direct link under the **Host** column.

2. Switch to the **Remote APIs** tab, find **SAP Commerce Cloud - Events** and select it.

3. Select the `order.deliverysent.v1` event type in **Event Topics** drop-down list. In the details of the printed event, change **orderCode** to `123456789` and select **Send Event**.

   The message appears on the UI confirming that the event was sent.

4. Call the microservice for the last time to verify is the event details were saved:

   ```bash
   curl -ik "https://$SERVICE_DOMAIN/orders"
   ```

   > **NOTE**: To get the domain of the microservice, run:
   >
   > ```bash
   > export SERVICE_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-service.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
   > ```

   You should see a response similar to the following:

   ```bash
   content-length: 2
   content-type: application/json;charset=UTF-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   server: istio-envoy
   vary: Origin
   x-envoy-upstream-service-time: 37

   [{"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"123456789","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```

5. Remove the [Pod](https://kubernetes.io/docs/concepts/workloads/pods/) created by the `orders-service` Deployment. Run this command and wait for the system to delete the Pod and start a new one:

      ```bash
      kubectl delete pod -n orders-service -l app=orders-service
      ```

6. Call the microservice again to check the storage:

      ```bash
      curl -ik "https://$SERVICE_DOMAIN/orders"
      ```

You will see that the order data was not removed this time. This proves that the details were saved in the Redis database instead of the default in-memory storage.
