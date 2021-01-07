---
title: Create an inline Function
type: Tutorials
---

This tutorial shows how you can create a simple "Hello World!" Function in Node.js 12. The Function's code and dependencies are defined as an inline code in the Function's **spec**.

> **TIP:** Serverless also allows you to store Function's code and dependencies as sources in a Git repository. To learn more, read how to [Create a Function from Git repository sources](#tutorials-create-a-function-from-git-repository-sources).

> **TIP:** As of 1.17, you can use Kyma CLI to create Functions, apply them on a cluster, and locally fetch the current state of your Function's cluster configuration after it was modified. To learn more, read how to [Use Kyma CLI to manage Functions](https://kyma-project.io/docs/cli/#tutorials-use-kyma-cli-to-manage-functions).

## Steps

Follows these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1.  Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    ```

2.  Create a Function CR that specifies the Function's logic:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      runtime: nodejs12
      source: |
        module.exports = {
          main: function(event, context) {
            return 'Hello World!'
          }
        }
    EOF
    ```

3.  Check if your Function was created successfully and all conditions are set to `True`:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

    You should get a result similar to the following example:

    ```bash
    NAME            CONFIGURED   BUILT     RUNNING   RUNTIME    VERSION   AGE
    test-function   True         True      True      nodejs12   1         96s
    ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. From the drop-down list in the top navigation panel, create a Namespace or select one.

2.  In the left navigation panel, go to **Workloads** > **Functions** and select **Create Function**.

3.  In the pop-up box, provide the Function's name, leave the default runtime `Node.js 12`, and select **Create** to confirm changes.

    The pop-up box closes and the message appears on the screen after a while, confirming that the Function was created successfully.

4.  In the Function details view that opens up automatically, enter the Function's code in the **Source** tab:

    ```
    module.exports = {
      main: function (event, context) {
      return 'Hello World!'
      }
    }
    ```

5.  Select **Save** to confirm changes.

    You will see the message confirming the changes were saved. Once deployed, the new Function should have the `RUNNING` status in the list of all Functions under the **Functions** view.

</details>
</div>
