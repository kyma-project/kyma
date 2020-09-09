---
title: Create a Function from Git Repository
type: Tutorials
---

This tutorial shows how you can create a sample Function from Git repository based on [orders service example](https://github.com/kyma-project/examples/tree/master/orders-service).

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

    (Optional) If you are willing to use a private repository, create a secret with credentials to the repository:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Secret
    metadata:
      name: git-creds-basic
      namespace: $NAMESPACE
    type: Opaque
    data:
      username: <USERNAME>
      password: <PASSWORD>
    EOF
    ```

    >**NOTE** To see other authorization methods, go to the [documentation]().

2. Create a GitRepository CR that specifies the Git repository metadata:

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
   
    >**NOTE** If you are using a private repository, add `auth` object with `type` and `secretName` fields to the spec. For details, see [documentation]().

3. Create a Function CR that specifies the Function's logic:

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

    >**NOTE** To see the function files, go to [this](https://github.com/kyma-project/examples/tree/master/orders-service/function) page. 

4. Check if your Function was created successfully and all conditions are set to `True`:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

    You should get a result similar to the following example:

    ```bash
    NAME                        CONFIGURED   BUILT   RUNNING   VERSION   AGE
    test-function               True         True    True      1         18m
    ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. Create a Namespace or select one from the drop-down list in the top navigation panel.

    (Optional) If you are willing to use a private repository, to authorize with basic authentication create a secret with credentials to the repository:

    1. Save yaml file with secret on your hard drive:
    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: git-creds-basic
    type: Opaque
    data:
      username: <USERNAME>
      password: <PASSWORD>
    ```
   >**NOTE** To see other possible authorization methods, go to the [documentation]().
   
   2. Go to the namespace view and click **Deploy new resource**.
   3. Select secret file and click **Deploy**.

2. Go to the **Functions** view in the left navigation panel and select **Repositories** tab.

3. Click **Connect Repository**, fill the `Url` field with `https://github.com/kyma-project/examples.git` value and click **Connect**.

    >**NOTE** If you want connect the private repository, change the `Authorization` field from `Public` to `Key` or `Basic` and fill the required fields. For details, see [documentation]().

4. Go to the **Functions** tab and click **Create Function**.

5. In the pop-up box, change `Source type` to `From Repository`, select created Repository's name, fill the `Reference` field with `master` and `Base directory` field with `orders-service/function` values and select **Create** to confirm changes.

    The pop-up box closes and the message appears on the screen after a while, confirming that the Function was created successfully.
    The new Function should have the `RUNNING` status in the list of all Functions under the **Functions** view.

    </details>
</div>
