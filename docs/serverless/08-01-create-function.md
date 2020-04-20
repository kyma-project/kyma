---
title: Create a function
type: Tutorials
---

This tutorial shows how you can create a simple "Hello World!" function.

## Steps

Follows these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    ```

2. Create a Function CR that specifies the function's logic:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      source: |
        module.exports = {
          main: function(event, context) {
            return 'Hello World!'
          }
        }
    EOF
    ```

3. Check if your function was created successfully and all conditions are set to `True`:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

    You should get a result similar to the following example:

    ```bash
    NAME                        CONFIGURED   BUILT   RUNNING   VERSION   AGE
    test-function                 True         True    True      1         18m
    ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. Create a Namespace or select one from the drop-down list in the top navigation panel.

2. Go to the **Functions** view in the left navigation panel and select **Create function**.

3. In the pop-up box, provide the function's name and select **Create** to confirm changes.

     The pop-up box closes and the message appears on the screen confirming that the function was created successfully.

4. In the function details view that opens up automatically, go to the **Code** tab and enter the function's code:

    ```
    module.exports = {
      main: function (event, context) {
      return 'Hello World!'
      }
    }
    ```

5. Select **Save** to confirm changes.

    You will see the message confirming the changes were saved. Once deployed, the new function should have the `RUNNING` status in the list of all functions under the **Functions** view.

    </details>
</div>
