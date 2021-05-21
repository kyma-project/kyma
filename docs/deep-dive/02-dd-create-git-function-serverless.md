---
title: Create a Git Function
type: Deep Dive
---

This tutorial shows how you can build a Function from code and dependencies stored in a Git repository, which is an alternative way to keeping the code in the Function CR. The tutorial is based on the Function from the [`orders service` example](https://github.com/kyma-project/examples/tree/main/orders-service). It describes steps required to fetch Function's source code and dependencies from a public Git repository that does not require any authentication method. However, it also provides additional guidance on how to secure it if you are using a private repository.

>**NOTE:** To learn more about Git repository sources for Functions and different ways of securing your repository, read about the [Git source type](#details-git-source-type).

## Steps

Follows these steps:

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
      username: {USERNAME}
      password: {PASSWORD_OR_TOKEN}
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
      key: {SSH_KEY}
    EOF
    ```

    >**NOTE:** Read more about the [supported authentication methods](#details-git-source-type).

3. Create a [GitRepository CR](#custom-resource-git-repository) that specifies the Git repository metadata:

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
      runtime: nodejs12
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
    test-function   True         True      True      nodejs12   1         96s
    ```

    </details>
    <details>
    <summary label="busola-ui">
    Busola UI
    </summary>

>**NOTE:** Busola is not installed by default. Follow the [instructions](https://github.com/kyma-project/busola/blob/main/README.md) to install it with npx.

1. Create a Namespace or select one from the drop-down list in the top navigation panel.

2. Create a Secret (optional).

  If you use a secured repository, you must first create a Secret with either basic (username and password or token) or SSH key authentication to this repository in the same Namespace as the Function. To do that, follow these sub-steps:

      2a. Open your Namespace view. In the left navigation panel, go to **Configuration** > **Secrets** and select the **Create Secret** button.

      2b. In the **Metadata** tab, enter the Secret name under **Name**.

      2c. In the **Data** tab, select **Add data entry** and enter these key-value pairs with credentials:

      - Basic authentication:

        ```bash
        username: {USERNAME}
        password: {PASSWORD_OR_TOKEN}
        ```

      - SSH key:

        ```bash
        key: {SSH_KEY}
        ```

      >**NOTE:** Read more about the [supported authentication methods](#details-git-source-type). 

     2d. Confirm by selecting **Create**.

3. In the left navigation panel, go to **Workloads** > **Functions** and select **Connected repositories**.

4. Select **Connect Repository**, fill in the **URL** field with `https://github.com/kyma-project/examples.git`, and confirm by selecting **Connect**.

    >**NOTE:** If you want to connect a secured repository, change the **Authorization** field from `Public` to `Basic` or `SSH key` and fill in the required fields.

5. Go back to the **Functions** view and select **Create Function**.

6. In the pop-up box, change **Source Type** to `Git repository`. Select the created repository's name and fill in the **Reference** field with `main` and the **Base Directory** field with `orders-service/function`. Select **Create** to confirm changes.

    The pop-up box closes and the message appears on the screen after a while, confirming that the Function was created.
    Make sure that the new Function has the `RUNNING` status.

    </details>
</div>
