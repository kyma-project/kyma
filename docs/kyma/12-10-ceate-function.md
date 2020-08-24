---
title: Create a Function
type: Getting Started
---

This tutorial shows how you can create a simple Function.

## Related Kyma components

This guide demonstrates how [Serverless](/components/event-mesh/) works in Kyma. It allows you to build, run, and manage serverless applications called Functions. You can bind them to other services, subscribe business events from external solutions to them, and trigger the Function's logic upon receiving a given event type.

## Steps

Follows these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Apply the Function CR that specifies the Function's logic:

```bash
kubectl apply -f https://raw.githubusercontent.com/kyma-project/examples/master/orders-service/deployment/orders-function.yaml
```

2. Check if the Function was created and all its conditions are set to `True`:

    ```bash
    kubectl get functions orders-function -n orders-service
    ```

    You should get a result similar to the following example:

    ```bash
    NAME                CONFIGURED   BUILT   RUNNING   VERSION   AGE
    orders-function     True         True    True      1         18m
    ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. Navigate to the `orders-service` Namespace view in the Console UI from the drop-down list in the top navigation panel.

2. Go to the **Functions** view under the **Development** section in the left navigation panel and select **Create Function**.

3. In the pop-up box, provide the `orders-function` name, add `app=orders-function` and `example=orders-function` labels, and select **Create** to confirm the changes.

>**TIP:** Separate multiple Function labels in the Console UI with commas.

     The pop-up box closes and the message appears on the screen after a while, confirming that the Function was created.

4. In the **Source** tab of the Function details view that opens up automatically, enter the Function's code from the [`orders-function.js`](https://raw.githubusercontent.com/kyma-project/examples/master/orders-service/deployment/orders-function.js) file.

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

  You will see the message confirming the changes were saved. Once deployed, the new Function should have the `RUNNING` status.

    </details>
</div>
