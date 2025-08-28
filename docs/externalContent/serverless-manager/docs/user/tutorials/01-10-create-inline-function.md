# Create and Modify an Inline Function

This tutorial shows how you can create a simple "Hello World" Function in Node.js. The Function's code and dependencies are defined as an inline code in the Function's **spec**.

Serverless also allows you to store the Function's code and dependencies as sources in a Git repository. To learn more, read how to [Create a Git Function](01-11-create-git-function.md).
To learn more about Function's signature, `event` and `context` objects, and custom HTTP responses the Function returns, read [Function’s specification](../technical-reference/07-70-function-specification.md).

> [!NOTE]
> Read about [Istio sidecars in Kyma and why you want them](https://kyma-project.io/docs/kyma/latest/01-overview/service-mesh/smsh-03-istio-sidecars-in-kyma/). Then, check how to [enable automatic Istio sidecar proxy injection](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection/). For more details, see [Default Istio setup in Kyma](https://kyma-project.io/docs/kyma/latest/01-overview/service-mesh/smsh-02-default-istio-setup-in-kyma/).

## Prerequisites

* You have the [Serverless module added](https://kyma-project.io/#/02-get-started/01-quick-install).
* For the Kyma CLI scenario, you have Kyma CLI installed.

## Procedure

You can create a Function with Kyma dashboard, Kyma CLI, or kubectl:

<!-- tabs:start -->

#### Kyma Dashboard

1. Create a namespace or select one from the drop-down list in the navigation panel.

2. Go to **Workloads** > **Functions** and choose **Create**.

3. In the dialog box, provide the Function's name.

4. Choose `JavaScript` from the **Language** dropdown.

5. Paste the following code snippet in the **Source** section, and choose **Create**.

   ```js
   module.exports = {
     main: async function (event, context) {
       const message =
         `Hello World` +
         ` from the Kyma Function ${context['function-name']}` +
         ` running on ${context.runtime}!`;
       console.log(message);
       return message;
     },
   };
   ```

Wait for the **Status** field to change into `RUNNING`, confirming that the Function was created successfully.

If you decide to modify your Function, choose **Edit**, make the changes, and choose the **Save** button. If successful, the message at the bottom of the screen confirms that the Function was updated.

#### Kyma CLI

1. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export KUBECONFIG={PATH_TO_YOUR_KUBECONFIG}
    ```

2. Create your local development workspace.

    a. Create a new folder to keep the Function's code and dependencies in one place:

    ```bash
    mkdir {FOLDER_NAME}
    cd {FOLDER_NAME}
    ```

    b. Create initial scaffolding for the Function:

    ```bash
    kyma alpha function init
    ```

3. Code and configure.

    Open the workspace in your favorite IDE. If you have Visual Studio Code installed, run the following command from the terminal in your workspace folder:

    ```bash
    code .
    ```

    It's time to inspect the code and the `handler.js` and the `package.json` files. Feel free to adjust the "Hello World" sample code.

4. Deploy and verify.

    a. Call the `create` command from the workspace folder. It builds the container and runs it on the Kyma runtime pointed by your current KUBECONFIG file:

      ```bash
      kyma alpha function create ${NAME} --namespace ${NAMESPACE} --runtime nodejs22 --source handler.js --dependencies package.json
      ```

    b. Check if your Function was created successfully:

      ```bash
      kyma alpha function get ${NAME} --namespace ${NAMESPACE}
      ```

    You should get a result similar to this example:

    ```bash
    NAME       CONFIGURED   BUILT   RUNNING   RUNTIME    GENERATION
    nodejs22   True         True    True      nodejs22   1
    ```

#### kubectl

1. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export KUBECONFIG={PATH_TO_YOUR_KUBECONFIG}
    ```

2. Create a Function CR that specifies the Function's logic:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: serverless.kyma-project.io/v1alpha2
   kind: Function
   metadata:
     name: $NAME
     namespace: $NAMESPACE
   spec:
     runtime: nodejs20
     source:
       inline:
         source: |
           module.exports = {
             main: function(event, context) {
               return 'Hello World!'
             }
           }
   EOF
   ```

3. Check if your Function was created successfully and all conditions are set to `True`:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

    You should get a result similar to this example:

    ```bash
    NAME            CONFIGURED   BUILT     RUNNING   RUNTIME    VERSION   AGE
    test-function   True         True      True      nodejs20   1         96s
    ```

<!-- tabs:end -->
