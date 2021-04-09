---
title: Create a Function from Git repository sources
type: Tutorials
---

This tutorial shows how you can build a Function from code and dependencies stored in a Git repository, which is an alternative way to keeping the code in the Function CR. The tutorial is based on the Function from the [`orders service` example](https://github.com/kyma-project/examples/tree/main/orders-service). It describes steps required to fetch Function's source code and dependencies from a public Git repository that does not require any authentication method. However, it also provides additional guidance on how to secure it if you are using a private repository.

> **NOTE:** This tutorial shows an alternative way of storing Function's code and dependencies. If you want to follow the whole end-to-end flow described for Functions in the Serverless tutorials, [create an inline Function](#tutorials-create-an-inline-function) instead. To learn more about Git repository sources for Functions and different ways of securing your repository, read about the [Git source type](#details-git-source-type).

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

1. Create a Namespace or select one from the drop-down list in the top navigation panel.

2. Create a Secret (optional).

  If you use a secured repository, you must first create a Secret with either basic (username and password or token) or SSH key authentication to this repository in the same Namespace as the Function. To do that, follow these sub-steps:

  - On your machine, create this YAML file with one of these Secret definitions:

    - Basic authentication:

      ```yaml
      apiVersion: v1
      kind: Secret
      metadata:
        name: git-creds-basic
        namespace: {NAMESPACE}
      type: Opaque
      data:
        username: {USERNAME}
        password: {PASSWORD_OR_TOKEN}
      ```

    - SSH key:


      ```yaml
      apiVersion: v1
      kind: Secret
      metadata:
        name: git-creds-key
        namespace: {NAMESPACE}
      type: Opaque
      data:
        key: {SSH_KEY}
      ```

    >**NOTE:** Read more about the [supported authentication methods](#details-git-source-type).

  - Go to your Namespace view and select **Deploy new workload** > **Upload YAML**.

  - Locate the YAML file with the Secret and select **Deploy**.

3. In the left navigation panel, go to **Workloads** > **Functions** and select the **Repositories** tab.

4. Select **Connect Repository**, fill in the **URL** field with `https://github.com/kyma-project/examples.git`, and confirm by selecting **Connect**.

    >**NOTE:** If you want to connect a secured repository, change the **Authorization** field from `Public` to `Basic` or `SSH key` and fill in the required fields.

5. Go to the **Functions** tab and select **Create Function**.

6. In the pop-up box, change `Source Type` to `From Repository`. Select the created repository's name and fill in the `Reference` field with `main` and the `Base Directory` field with `orders-service/function`. Select **Create** to confirm changes.

    The pop-up box closes and the message appears on the screen after a while, confirming that the Function was created.
    Make sure that the new Function has the `RUNNING` status in the list of all Functions under the **Functions** view.

    </details>
</div>
