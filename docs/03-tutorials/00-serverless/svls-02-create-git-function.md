---
title: Create a Git Function
---

This tutorial shows how you can build a Function from code and dependencies stored in a Git repository, which is an alternative way to keeping the code in the Function CR. The tutorial is based on the Function from the [`orders service` example](https://github.com/kyma-project/examples/tree/main/orders-service). It describes steps required to fetch the Function's source code and dependencies from a public Git repository that does not need any authentication method. However, it also provides additional guidance on how to secure it if you are using a private repository.

>**NOTE:** To learn more about Git repository sources for Functions and different ways of securing your repository, read about the [Git source type](../../05-technical-reference/svls-04-git-source-type.md).

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

3. Create a [GitRepository CR](../../05-technical-reference/00-custom-resources/svls-02-gitrepository.md) that specifies the Git repository metadata:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: GitRepository
    metadata:
      name: $GIT_FUNCTION
      namespace: $NAMESPACE
    spec:
      url: "https://github.com/kyma-project/examples.git"
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
4. Create a Function CR that specifies the Function's logic and points to the directory with code and dependencies in the given repository.

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $GIT_FUNCTION
      namespace: $NAMESPACE
    spec:
      type: git
      runtime: nodejs14
      source: $GIT_FUNCTION
      reference: main
      baseDir: orders-service/function
    EOF
    ```

    >**NOTE:** See this [Function's code and dependencies](https://github.com/kyma-project/examples/tree/main/orders-service/function).

5. Check if your Function was created and all conditions are set to `True`:

    ```bash
    kubectl get functions $GIT_FUNCTION -n $NAMESPACE
    ```

    You should get a result similar to this example:

    ```bash
    NAME            CONFIGURED   BUILT     RUNNING   RUNTIME    VERSION   AGE
    test-function   True         True      True      nodejs14   1         96s
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

3. To connect the repository, go to **Workloads** > **Functions** > **Connected repositories**.

4. Connect your repository, with `https://github.com/kyma-project/examples.git` as repository URL.

    >**NOTE:** If you want to connect a secured repository instead of a public one, select authorization method `Basic` or `SSH key` and fill in the required fields.

5. Go back to the **Functions** view and select **Create Function**.

6. Under **Advanced**, change the source type to `Git repository` and select the created repository's name. As reference, enter `main`, and as base directory, `orders-service/function`.

    After a while, a message confirms that the Function has been created.
    Make sure that the new Function has the `RUNNING` status.

    </details>
</div>
