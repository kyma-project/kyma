---
title: Create an inline Function
type: Deep Dive
---

This tutorial shows how you can create a simple "Hello World!" Function in Node.js 12. The Function's code and dependencies are defined as an inline code in the Function's **spec**.

> **TIP:** Serverless also allows you to store Function's code and dependencies as sources in a Git repository. To learn more, read how to [Create a Function from Git repository sources](#tutorials-create-a-function-from-git-repository-sources).

> **TIP:** Read about [Function’s specification](#details-function-s-specification) if you are interested in its signature, `event` and `context` objects, and custom HTTP responses the Function returns.

## Steps

Follows these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="CLI">
  kyma CLI
  </summary>

1.  Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    ```

2.  Create a local development workspace

    Create a new folder to keep function's code and configuration in one place.

    ```bash
    mkdir my-function // Call it whatever you want
    cd my-function
    ```

    Create an inital scaffolding using dedicated CLI command.

    ```bash
    kyma init function --name $NAME --namespace $NAMESPACE
    ```

    Open

3.  Code & configure.

    Open the workspace in your favorite IDE. If you have VS Code installed just call the following from the terminal in your workspace folder

    ```bash
    code .
    ```

    Its time to inspect the code & config.yaml. Feel free to adjust the "hello world" sample code

4.  Deploy

    Call the apply command from the workspace folder. It will build the container and run it on the kyma runtime pointed by your current kubeconfig.

    ```bash
    kyma apply function
    ```

    Check if your Function was created successfully

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

    You should get a result similar to the following example:

    ```bash
    NAME            CONFIGURED   BUILT     RUNNING   RUNTIME    VERSION   AGE
    test-function   True         True      True      nodejs12   1         96s
    ```

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

4.  From the drop-down list in the top navigation panel, create a Namespace or select one.

5.  In the left navigation panel, go to **Workloads** > **Functions** and select **Create Function**.

6.  In the pop-up box, provide the Function's name, leave the default runtime `Node.js 12`, and select **Create** to confirm changes.

    The pop-up box closes and the message appears on the screen after a while, confirming that the Function was created successfully.

7.  In the Function details view that opens up automatically, enter the Function's code in the **Source** tab:

    ```
    module.exports = {
      main: function (event, context) {
      return 'Hello World!'
      }
    }
    ```

8.  Select **Save** to confirm changes.

    You will see the message confirming the changes were saved. Once deployed, the new Function should have the `RUNNING` status in the list of all Functions under the **Functions** view.

</details>
</div>
