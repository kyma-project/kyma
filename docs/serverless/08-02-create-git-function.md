---
title: Create a Function from Git Repository
type: Tutorials
---

This tutorial shows how you can create a simple "Hello World!" Function from Git repository.

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

2. Create a GitRepository CR that specifies the Git repository metadata:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: GitRepository
    metadata:
      name: $NAME
    spec:
      url: "https://github.com/pPrecel/public-gitops"
    EOF
    ```
   
>**NOTE** Auth possibility TODO

3. Create a Function CR that specifies the Function's logic:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
    spec:
      type: git
      runtime: nodejs12
      source: $NAME
      reference: master
      baseDir: js-handler
    EOF
    ```

    >**NOTE** To see full spec, go to the xxx page 

3. Check if your Function was created successfully and all conditions are set to `True`:

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

2. Go to the **Functions** view in the left navigation panel and select **Repositories** tab.

3. Click **Connect Repository**,  fill in required fields and click **Connect**.

    >**NOTE** Auth possibility TODO

4. Go to the **Functions** tab and click **Create Function**.

3. In the pop-up box, change `Source type` to `From Repository`, fill in required fields and select **Create** to confirm changes.

    The pop-up box closes and the message appears on the screen after a while, confirming that the Function was created successfully.
    The new Function should have the `RUNNING` status in the list of all Functions under the **Functions** view.

    </details>
</div>
