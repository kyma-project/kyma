---
title: Create a lambda
type: Tutorials
---

This tutorial shows how you can create a simple "Hello World!" lambda.

## Steps

Follows these steps:

<div tabs name="steps" group="create-lambda">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Export these variables:

    ```bash
    export NAME={LAMBDA_NAME}
    export NAMESPACE={LAMBDA_NAMESPACE}
    ```

2. Create a Function CR that specifies the lambda's logic and defines a runtime on which it should run:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      functionContentType: plaintext
      runtime: nodejs8
      function: |
        module.exports = {
          main: function(event, context) {
            return 'Hello World!'
          }
        }
    EOF    
    ```

3. Check if your lambda was created successfully and has the `Running` status:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE -o=jsonpath='{.status.condition}'
    ```

    </details>
<details>
<summary label="console-ui">
Console UI
</summary>

> **NOTE:** Serverless v2 is an experimental feature, and it is not enabled by default in the Console UI. To use its **Functions [preview]** view, enable **Experimental functionalities** in the **General Settings** view before you follow the steps.

1. Select a Namespace from the drop-down list in the top navigation panel or create a new one.
2. Go to the **Functions [preview]** view at the bottom of the left navigation panel and select **Create lambda**.
3. In the pop-up box, provide lambda's name and select **Create** to confirm changes.

The pop-up box closes and you will see the `Lambda created successfully` message.

4. In the lambda details view that opens up automatically, go to the **Code** tab and enter the lambda's code:

```
module.exports = {
  main: function (event, context) {
  return 'Hello World!'
  }
}
```

5. Select **Save** to confirm changes.

You will get the `Lambda created successfully` message confirming the changes were saved. Once deployed, the new lambda should have the `RUNNING` status in the list of all lambdas under the **Functions [preview]** view.


    </details>
