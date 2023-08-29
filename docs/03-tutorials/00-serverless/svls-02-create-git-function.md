---
title: Create a Git Function
---

This tutorial shows how you can build a Function from code and dependencies stored in a Git repository, which is an alternative way to keeping the code in the Function CR. The tutorial is based on the Function from the [`orders service` example](https://github.com/kyma-project/examples/tree/main/orders-service). It describes steps required to fetch the Function's source code and dependencies from a public Git repository that does not need any authentication method. However, it also provides additional guidance on how to secure it if you are using a private repository.

>**NOTE:** To learn more about Git repository sources for Functions and different ways of securing your repository, read about the [Git source type](../../05-technical-reference/svls-04-git-source-type.md).

>**NOTE:** Read about [Istio sidecars in Kyma and why you want them](/istio-operator/user/00-overview/00-30-overview-istio-sidecars). Then, check how to [enable automatic Istio sidecar proxy injection](/istio-operator/user/02-operation-guides/operations/02-20-enable-sidecar-injection). For more details, see [Default Istio setup in Kyma](/istio-operator/user/00-overview/00-40-overview-istio-setup).

## Steps

Follow these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Export these variables:

    ```bash
    export GIT_FUNCTION={GIT_FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    ```

2. Create a Secret (optional).

    If you use a secured repository, you must first create a Secret for one of these authentication methods:

    - Basic authentication (username and password or token) to this repository in the same Namespace as the Function:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Secret
    metadata:
      name: git-creds-basic
      namespace: $NAMESPACE
    type: Opaque
    data:
      username: {BASE64_ENCODED_USERNAME}
      password: {BASE64_ENCODED_PASSWORD_OR_TOKEN}
    EOF
    ```

    - SSH key:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Secret
    metadata:
      name: git-creds-key
      namespace: $NAMESPACE
    type: Opaque
    data:
      key: {BASE64_ENCODED_PRIVATE_SSH_KEY}
    EOF
    ```

    >**NOTE:** Read more about the [supported authentication methods](../../05-technical-reference/svls-04-git-source-type.md).

3. Create a Function CR that specifies the Function's logic and points to the directory with code and dependencies in the given repository. It also specifies the Git repository metadata:

   ```yaml
   cat <<EOF | kubectl apply -f -
   apiVersion: serverless.kyma-project.io/v1alpha2
   kind: Function
   metadata:
     name: $GIT_FUNCTION
     namespace: $NAMESPACE
   spec:
     runtime: nodejs18
     source:
       gitRepository:
         baseDir: orders-service/function
         reference: main
         url: https://github.com/kyma-project/examples.git
   EOF
   ```

    >**NOTE:** If you use a secured repository, add the **auth** object with the adequate **type** and **secretName** fields to the spec:

    ```yaml
    spec:
      ...
      auth:
        type: # "basic" or "key"
        secretName: # "git-creds-basic" or "git-creds-key"
    ```
   
    >**NOTE:** To avoid performance degradation caused by large Git repositories and large monorepos, Function Controller implements a configurable backoff period for the source checkout based on `APP_FUNCTION_REQUEUE_DURATION`. This behavior can be disabled, allowing the controller to perform the source checkout with every reconciliation loop by marking the Function CR with the annotation `serverless.kyma-project.io/continuousGitCheckout: true`

    >**NOTE:** See this [Function's code and dependencies](https://github.com/kyma-project/examples/tree/main/orders-service/function).

4. Check if your Function was created and all conditions are set to `True`:

    ```bash
    kubectl get functions $GIT_FUNCTION -n $NAMESPACE
    ```

    You should get a result similar to this example:

    ```bash
    NAME            CONFIGURED   BUILT     RUNNING   RUNTIME    VERSION   AGE
    test-function   True         True      True      nodejs18   1         96s
    ```

    </details>
    <details>
    <summary label="busola-ui">
    Kyma Dashboard
    </summary>

>**NOTE:** Kyma Dashboard uses Busola, which is not installed by default. Follow the [instructions](https://github.com/kyma-project/busola#installation) to install it.

1. Create a Namespace or select one from the drop-down list in the top navigation panel.

2. Create a Secret (optional).

    If you use a secured repository, you must first create a Secret with either basic (username and password or token) or SSH key authentication to this repository in the same Namespace as the Function. To do that, follow these sub-steps:

    - Open your Namespace view. In the left navigation panel, go to **Configuration** > **Secrets** and select the **Create Secret** button.

    - Open the **Advanced** view and enter the Secret name and type.

    - Select **Add data entry** and enter these key-value pairs with credentials:

        - Basic authentication: `username: {USERNAME}` and `password: {PASSWORD_OR_TOKEN}``

        - SSH key: `key: {SSH_KEY}`

        >**NOTE:** Read more about the [supported authentication methods](../../05-technical-reference/svls-04-git-source-type.md).

    - Confirm by selecting **Create**.

3. To connect the repository, go to **Workloads** > **Functions** > **Create Function**.

4. Provide or generate the Function's name. 

4. Go to **Advanced**, change **Source Type** from **Inline** to **Git Repository**.

5. Click on the **Git Repository** section and enter `https://github.com/kyma-project/examples.git` as repository **URL**, `orders-service/function` as **Base Dir**,  and `main` as **Reference**.

    > **NOTE:** If you want to connect a secured repository instead of a public one, toggle the **Auth** switch. In the **Auth** section choose **Secret** from the list and choose the preffered type.
6. Click **Create**.

    After a while, a message confirms that the Function has been created.
    Make sure that the new Function has the `RUNNING` status.

    </details>
</div>
