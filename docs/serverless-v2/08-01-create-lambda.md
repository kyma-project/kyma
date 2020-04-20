---
title: Create a lambda
type: Tutorials
---

This tutorial shows how you can create a simple "Hello World!" lambda.

## Steps

Follows these steps:

<div tabs name="steps" group="create-lambda">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Export these variables:

    ```bash
    export NAME={LAMBDA_NAME}
    export NAMESPACE={LAMBDA_NAMESPACE}
    ```

2. Create a Function CR that specifies the lambda's logic:

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

3. Check if your lambda was created successfully and all conditions are set to `True`:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

    You should get a result similar to the following example:

    ```bash
    NAME                        CONFIGURED   BUILT   RUNNING   VERSION   AGE
    test-lambda                 True         True    True      1         18m
    ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

> **NOTE:** Serverless v2 is an experimental feature, and it is not enabled by default in the Console UI. To use its **Functions [preview]** view, enable **Experimental functionalities** in the **General Settings** view before you follow the steps. Refresh the page after enabling this option.

1. Create a Namespace or select one from the drop-down list in the top navigation panel.

2. Go to the **Functions [preview]** view at the bottom of the left navigation panel and select **Create lambda**.

3. In the pop-up box, provide the lambda's name and select **Create** to confirm changes.

     The pop-up box closes and the `Lambda created successfully` message appears.

4. In the lambda details view that opens up automatically, go to the **Code** tab and enter the lambda's code:

    ```
    module.exports = {
      main: function (event, context) {
      return 'Hello World!'
      }
    }
    ```

5. Select **Save** to confirm changes.

    The `Lambda {NAME} updated successfully` message appears confirming the changes were saved. Once deployed, the new lambda should have the `RUNNING` status in the list of all lambdas under the **Functions [preview]** view.

    </details>
</div>
