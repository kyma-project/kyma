# Create a Git Function

This tutorial shows how you can build a Function from code and dependencies stored in a Git repository, which is an alternative way to keeping the code in the Function CR. The tutorial is based on the Function from the [`redis rest` example](https://github.com/kyma-project/serverless/tree/main/examples/redis-rest). It describes steps required to fetch the Function's source code and dependencies from a public Git repository that does not need any authentication method. However, it also provides additional guidance on how to secure it if you are using a private repository.

To learn more about Git repository sources for Functions and different ways of securing your repository, read about the [Git source type](../technical-reference/07-40-git-source-type.md).

> [!NOTE]
> Read about the [purpose and benefits of Istio sidecar proxies](https://kyma-project.io/#/istio/user/00-00-istio-sidecar-proxies?id=purpose-and-benefits-of-istio-sidecar-proxies). Then, check how to [enable Istio sidecar proxy injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection). For more details, see [Default Istio Configuration](https://kyma-project.io/#/istio/user/00-15-overview-istio-setup?id=default-istio-configuration).

## Prerequisites

* You have the Serverless module added.

## Procedure

You can create a Function either with kubectl or Kyma dashboard:

<Tabs>
<Tab name="Kyma Dashboard">

> [!NOTE]
> Kyma dashboard uses Busola, which is not installed by default. Follow the [installation instructions](https://github.com/kyma-project/busola/blob/main/docs/install-kyma-dashboard-manually.md).

1. Create a namespace or select one from the drop-down list in the top navigation panel.

2. Create a Secret (optional).

    If you use a secured repository, you must first create a Secret with either basic (username and password or token) or SSH key authentication to this repository in the same namespace as the Function. To do that, follow these sub-steps:

    1. Open your namespace view. In the left navigation panel, go to **Configuration** > **Secrets** and choose **Create**.

    2. Enter the Secret name and type.

    3. Under **Data**, enter these key-value pairs with credentials:

        - Basic authentication: `username: {USERNAME}` and `password: {PASSWORD_OR_TOKEN}`

        - SSH key: `key: {SSH_KEY}`

        > [!NOTE]
        > Read more about the [supported authentication methods](../technical-reference/07-40-git-source-type.md).

    4. Confirm by selecting **Create**.

3. Go to **Workloads** > **Functions** > **Create**.

4. Provide the Function's name.

5. To connect the repository, change **Source Type** from **Inline** to **Git Repository**.

6. Choose `JavaScript` from the **Language** dropdown and select the proper runtime.

7. Click on the **Git Repository** section and enter the following values:
   - Repository **URL**: `https://github.com/kyma-project/serverless.git`
   - **Base Dir**:`examples/redis-rest`
   - **Reference**:`main`

    > [!NOTE]
    > If you want to connect a secured repository instead of a public one, toggle the **Auth** switch. In the **Auth** section, choose **Secret** from the list and choose the preferred type.

8. Click **Create**.

    After a while, a message confirms that the Function has been created.
    Make sure that the new Function has the `RUNNING` status.
</Tab>
<Tab name="kubectl">

1. Export these variables:

    ```bash
    export GIT_FUNCTION={GIT_FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export KUBECONFIG={PATH_TO_YOUR_KUBECONFIG}
    ```

2. Create a Secret (optional).

    If you use a secured repository, follow the sub-steps for the basic or SSH key authentication:

    - Basic authentication (username and password or token)
  
    1. Generate a [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-personal-access-token-classic) and copy it.
    2. Create a Secret containing your username and the generated token. You must create the Secret in the same namespace as the Function.

       ```bash
       kubectl -n $NAMESPACE create secret generic git-creds-basic --from-literal=username={GITHUB_USERNAME} --from-literal=password={GENERATED_PERSONAL_TOKEN}
       ```

    - SSH key:

    1. Generate a new SSH key pair (private and public). Follow [this tutorial](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent) to learn how to do it. Alternatively, you can use the existing pair.
    2. Install the generated private key in Kyma, as a Kubernetes Secret located in the same namespace as your Function.

       ```bash
       kubectl -n $NAMESPACE create secret generic git-creds-ssh --from-file=key={PATH_TO_THE_FILE_WITH_PRIVATE_KEY}
       ```

    3. Configure the public key in GitHub. Follow the steps described in [this tutorial](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/adding-a-new-ssh-key-to-your-github-account).

    > [!NOTE]
    > Read more about the [supported authentication methods](../technical-reference/07-40-git-source-type.md).

3. Create a Function CR that specifies the Function's logic and points to the directory with code and dependencies in the given repository. It also specifies the Git repository metadata:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: serverless.kyma-project.io/v1alpha2
   kind: Function
   metadata:
     name: $GIT_FUNCTION
     namespace: $NAMESPACE
   spec:
     runtime: nodejs20
     source:
       gitRepository:
         baseDir: examples/redis-rest/src
         reference: main
         url: https://github.com/kyma-project/serverless.git
   EOF
   ```

    > [!NOTE]
    > If you use a secured repository, add the **auth** object with the adequate **type** and **secretName** fields to the spec under **gitRepository**:

    ```yaml
    gitRepository:
      ...
      auth:
        type: # "basic" or "key"
        secretName: # "git-creds-basic" or "git-creds-ssh"
    ```

    If you use the `key` type authentication, you must use the SSH URL format to configure the Function URL:

    ```yaml
    gitRepository:
      ...
      url: git@github.com/<your-github-handle>/<repo-name>.git
      auth:
        type: key
        secretName: "git-creds-ssh"
    ```  

    > [!NOTE]
    > To avoid performance degradation caused by large Git repositories and large monorepos, [Function Controller](../resources/06-10-function-cr.md#related-resources-and-components) implements a configurable backoff period for the source checkout based on `APP_FUNCTION_REQUEUE_DURATION`. If you want to allow the controller to perform the source checkout with every reconciliation loop, disable the backoff period by marking the Function CR with the annotation `serverless.kyma-project.io/continuousGitCheckout: true`

    > [!NOTE]
    > See this [Function's code and dependencies](https://github.com/kyma-project/serverless/tree/main/examples/redis-rest).

4. Check if your Function was created and all conditions are set to `True`:

    ```bash
    kubectl get functions $GIT_FUNCTION -n $NAMESPACE
    ```

    You should get a result similar to this example:

    ```bash
    NAME            CONFIGURED   BUILT     RUNNING   RUNTIME    VERSION   AGE
    test-function   True         True      True      nodejs20   1         96s
    ```
</Tab>
</Tabs>
