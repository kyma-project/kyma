---
title: Create an inline Function
type: Deep Dive
---

This tutorial shows how you can create a simple "Hello World!" Function in Node.js 12. The Function's code and dependencies are defined as an inline code in the Function's **spec**.

> **TIP:** Serverless also allows you to store Function's code and dependencies as sources in a Git repository. To learn more, read how to [Create a Function from Git repository sources](#tutorials-create-a-function-from-git-repository-sources).

> **TIP:** Read about [Function’s specification](#details-function-s-specification) if you are interested in its signature, `event` and `context` objects, and custom HTTP responses the Function returns.

## Steps

Follow these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="CLI">
  Kyma CLI
  </summary>

1.  Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    ```

2.  Create your local development workspace.

    a. Create a new folder to keep Function's code and configuration in one place.

    ```bash
    mkdir {FOLDER_NAME}
    cd my-function
    ```

    b. Create initial scaffolding for the Function using the dedicated CLI command.

    ```bash
    kyma init function --name $NAME --namespace $NAMESPACE
    ```

3.  Code and configure.

    Open the workspace in your favorite IDE. If you have Visual Stdio Code installed, run the following command from the terminal in your workspace folder:

    ```bash
    code .
    ```

    It's time to inspect the code and the `config.yaml` file. Feel free to adjust the "Hello World" sample code.

4.  Deploy and verify.

    a. Call the `apply` command from the workspace folder. It will build the container and run it on the Kyma runtime pointed by your current KUBECONFIG file.

      ```bash
      kyma apply function
      ```

    b. Check if your Function was created successfully.

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
