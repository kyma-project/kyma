---
title: Create a Function from Git repository sources
type: Tutorials
---

Create a sample Function from code and dependencies stored in a Git repository. The tutorial is based on the Function from the [`orders service` example](https://github.com/kyma-project/examples/tree/master/orders-service). It shows how you can fetch Function's source code and dependencies from a public Git repository that does not require any authentication method.

> **NOTE:** Read more about [Git source type](#details-git-source-type) and different ways of securing your repository.

## Steps

Follows these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    ```

    If you use a secured repository, you must first create a Secret with basic authentication to this repository in the same Namespace as the Function:

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
      password: {PASSWORD}
    EOF
    ```

    >**NOTE:** Read also about other [supported authentication methods](#details-git-source-type).

2. Create a [GitRepository CR](#custom-resource-git-repository) that specifies the Git repository metadata:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: GitRepository
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      url: "https://github.com/kyma-project/examples.git"
    EOF
    ```

    >**NOTE:** If you use a secured repository, add the **auth** object with the **type** and **secretName** fields to the spec.

3. Create a Function CR that specifies the Function's logic and points to the directory with code and dependencies in the given repository.

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      type: git
      runtime: nodejs12
      source: $NAME
      reference: master
      baseDir: orders-service/function
    EOF
    ```

    >**NOTE:** See this [Function's code and dependencies](https://github.com/kyma-project/examples/tree/master/orders-service/function).

4. Check if your Function was created and all conditions are set to `True`:

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

1. Create a Namespace or select one from the drop-down list in the top navigation panel.

    If you use a secured repository, you must first create a Secret with basic authentication to this repository in the same Namespace as the Function. To do that, follow these sub-steps:

    - On your machine, create this YAML file with the Secret definition:

    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: git-creds-basic
    type: Opaque
    data:
      username: {USERNAME}
      password: {PASSWORD}
    ```

    >**NOTE:** Read also about other [supported authentication methods](#details-git-source-type).

    - Go to your Namespace view and select **Deploy new resource**.

    - Locate the YAML file with the Secret and select **Deploy**.


2. Go to the **Functions** view in the left navigation panel and select the **Repositories** tab.

3. Select **Connect Repository**, fill in the **Url** field with `https://github.com/kyma-project/examples.git`, and confirm by selecting **Connect**.

    >**NOTE:** If you want to connect the secured repository with basic authentication, change the **Authorization** field from `Public` to `Basic` and fill in the required fields.

4. Go to the **Functions** tab and select **Create Function**.

5. In the pop-up box, change `Source Type` to `From Repository`, select created Repository's name, fill in the `Reference` field with `master` and `Base Directory` field with `orders-service/function` values, and select **Create** to confirm changes.

    The pop-up box closes and the message appears on the screen after a while, confirming that the Function was created.
    Make sure that new Function has the `RUNNING` status in the list of all Functions under the **Functions** view.

    </details>
</div>
