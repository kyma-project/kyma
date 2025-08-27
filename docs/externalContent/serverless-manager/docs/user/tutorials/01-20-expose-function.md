# Expose a Function Using the APIRule Custom Resource

This tutorial shows how you can expose your Function to access it outside the cluster, through an HTTP proxy. To expose it, use an [APIRule custom resource (CR)](https://kyma-project.io/docs/kyma/latest/05-technical-reference/00-custom-resources/apix-01-apirule/). Function Controller reacts to an instance of the APIRule CR and, based on its details, it creates an Istio VirtualService and Oathkeeper Access Rules that specify your permissions for the exposed Function.

When you complete this tutorial, you get a Function that:

- Uses the `noAuth` access strategy, allowing access on an unsecured endpoint.
- Accepts the `GET`, `POST`, `PUT`, and `DELETE` methods.

To learn more about securing your Function, see the tutorial [Expose and secure a workload with JWT](https://kyma-project.io/#/api-gateway/user/tutorials/01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt).

Read also about [Function’s specification](../technical-reference/07-70-function-specification.md) if you are interested in its signature, `event` and `context` objects, and custom HTTP responses the Function returns.

## Prerequisites

- You have an [existing Function](01-10-create-inline-function.md).
- You have the [Istio, API Gateway, and Serverless modules added](https://kyma-project.io/#/02-get-started/01-quick-install).
- For the Kyma CLI scenario, you have Kyma CLI installed.

## Procedure

You can expose a Function using Kyma dashboard, Kyma CLI, or kubectl:

<Tabs>
<Tab name="Kyma Dashboard">

1. Select a namespace from the drop-down list in the navigation panel. Make sure the namespace includes the Function that you want to expose using the APIRule CR.

2. Go to **Discovery and Network** > **API Rules**, and choose **Create**.

3. Enter the following information:

    - The APIRule's **Name** matching the Function's name.

    > [!NOTE]
    > The APIRule CR can have a name different from that of the Function, but it is recommended that all related resources share a common name.

    - **Service Name** matching the Function's name.

    - **Host** to determine the host on which you want to expose your Function.

4. Edit the **Rules** section.
  - Select the methods `GET`, `POST`, `PUT`, and `DELETE`. 
  - Use the `No Auth` access strategy.

5. Select **Create** to confirm your changes.

6. To check if you can access the Function, copy the host link from the **General** section and paste it into your browser. If successful, the following message appears: `Hello World from the Kyma Function my-function running on nodejs20!`.
</Tab>
<Tab name="Kyma CLI">

> [!WARNING]
> This section is not yet compliant with Kyma CLI v3.
</Tab>
<Tab name="kubectl">

1. Run the following command to get the domain name of your Kyma cluster:

    ```bash
    kubectl get gateway -n kyma-system kyma-gateway \
        -o jsonpath='{.spec.servers[0].hosts[0]}'
    ```

2. Export the result without the leading `*.` as an environment variable:

    ```bash
    export DOMAIN={DOMAIN_NAME}

3. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export KUBECONFIG={PATH_TO_YOUR_KUBECONFIG}
    ```

    > [!NOTE]
    > The APIRule CR can have a name different from that of the Function, but it is recommended that all related resources share a common name.

4. Create an APIRule CR, which exposes your Function on port `80`.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      hosts:
      - $NAME
      service:
        name: $NAME
        namespace: $NAMESPACE
        port: 80
      gateway: kyma-system/kyma-gateway
      rules:
      - path: /*
        methods: ["GET", "POST", "PUT", "DELETE"]
        noAuth: true
    EOF
    ```

5. Check that the APIRule was created successfully and has the status `Ready`:

    ```bash
    kubectl get apirules $NAME -n $NAMESPACE -o=jsonpath='{.status.state}'
    ```

6. Access the Function's external address:

    ```bash
    curl https://$NAME.$DOMAIN
    ```

    If successful, the following mesage appears: `Hello World from the Kyma Function my-function running on nodejs20!`.
</Tab>
</Tabs>
