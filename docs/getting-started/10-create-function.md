---
title: Create a Function
type: Getting Started
---

Let's now repeat the microservice flow for the Function. This guide shows how you can create a simple Function (`orders-function`) with the same logic as the one in the microservice. In further guides, you will expose the Function, bind it to the Redis storage, and subscribe it to the `order.deliverysent.v1` event type from Commerce mock.

## Reference

This guide demonstrates how [Serverless](/components/event-mesh/) works in Kyma. It allows you to build, run, and manage serverless applications called Functions. You can bind them to other services, subscribe business events from external solutions to them, and trigger the Function's logic upon receiving a given event type.

## Steps

Follows these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Apply a [Function CR](/components/serverless/#custom-resource-function) that specifies the Function's logic:

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kyma-project/examples/master/orders-service/deployment/orders-function.yaml
  ```

2. Check that the Function was created and all its conditions are set to `True`:

    ```bash
    kubectl get functions orders-function -n orders-service
    ```

    Expect a response similar to this one:

    ```bash
    NAME                CONFIGURED   BUILT   RUNNING   VERSION   AGE
    orders-function     True         True    True      1         18m
    ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. From the drop-down list in the top navigation panel, select the `orders-service` Namespace.

2. Go to **Workloads** > **Functions** in the left navigation panel and select **Create Function**.

3. In the pop-up box, provide the `orders-function` name. Add the `app=orders-function` and `example=orders-function` labels, and select **Create** to confirm the changes.

  >**TIP:** Separate multiple Function labels in the Console UI with commas.

  The pop-up box will close and a message on the screen will confirm that the Function was created.

4. In the **Source** tab of the automatically opened Function details view, enter the Function's code from the [`handler.js`](https://raw.githubusercontent.com/kyma-project/examples/master/orders-service/function/handler.js) file.

5. In the **Dependencies** tab, enter:

  ```js
  {
    "name": "orders-function",
    "version": "1.0.0",
    "dependencies": {
      "redis": "3.0.2"
    }
  }
  ```

6. Select **Save** to confirm the changes.

  You will see a message confirming that the changes were saved. Once deployed, the new Function gets the status `RUNNING`.

    </details>
</div>
