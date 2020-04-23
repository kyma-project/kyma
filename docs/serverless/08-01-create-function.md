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

    > **CAUTION:** If you create a new Namespace, do not disable sidecar injection in it as Serverless requires Istio for other resources to communicate with functions correctly. Also, if you apply custom [LimitRanges](https://kyma-project.io/docs/#details-resource-quotas) for a new Namespace, they must be higher than the default values.

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

    > **CAUTION:** If you create a new Namespace, do not disable sidecar injection in it as Serverless requires Istio for other resources to communicate with functions correctly. Also, if you apply custom [LimitRanges](https://kyma-project.io/docs/#details-resource-quotas) for a new Namespace, they must be higher than the default values.

2. Go to the **Functions** view in the left navigation panel and select **Create Function**.

3. In the pop-up box, provide the function's name and select **Create** to confirm changes.

     The pop-up box closes and the message appears on the screen after a while, confirming that the function was created successfully.

4. In the function details view that opens up automatically, enter the function's code in the **Source** tab:

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
