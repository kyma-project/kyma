---
title: Expose a Function with an API Rule
---

This tutorial shows how you can expose your Function to access it outside the cluster, through an HTTP proxy. To expose it, use an APIRule custom resource (CR) managed by the in-house API Gateway Controller. This controller reacts to an instance of the APIRule CR and, based on its details, it creates an Istio VirtualService and Oathkeeper Access Rules that specify your permissions for the exposed Function.

When you complete this tutorial, you get a Function that:

- Is available on an unsecured endpoint (**handler** set to `noop` in the APIRule CR).
- Accepts the `GET`, `POST`, `PUT`, and `DELETE` methods.

>**NOTE:** To learn more about securing your Function, see the [Expose and secure a workload with OAuth2](../00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md) or [Expose and secure a workload with JWT](../00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md) tutorials.

>**TIP:** Read also about [Function’s specification](../../05-technical-reference/svls-07-function-specification.md) if you are interested in its signature, `event` and `context` objects, and custom HTTP responses the Function returns.

## Prerequisites

This tutorial is based on an existing Function. To create one, follow the [Create a Function](./svls-01-create-inline-function.md) tutorial.

>**NOTE:** Read about [Istio sidecars in Kyma and why you want them](../../01-overview/service-mesh/smsh-03-istio-sidecars-in-kyma.md). Then, check how to [enable automatic Istio sidecar proxy injection](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md). For more details, see [Default Istio setup in Kyma](../../01-overview/service-mesh/smsh-02-default-istio-setup-in-kyma.md).

## Steps

Follow these steps:

<div tabs name="steps" group="expose-function">
  <details>
  <summary label="cli">
  Kyma CLI
  </summary>

1. Export these variables:

      ```bash
      export DOMAIN={DOMAIN_NAME}
      export NAME={FUNCTION_NAME}
      export NAMESPACE={NAMESPACE_NAME}
      ```
   >**NOTE:** The Function takes the name from the Function CR name. The APIRule CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.
2. Download the latest configuration of the Function from the cluster. This way you will update the local `config.yaml` file with the Function's code.

  ```bash
  kyma sync function $NAME -n $NAMESPACE
  ```

3. Edit the local `config.yaml` file and add the **apiRules** schema for the Function at the end of the file:

  ```yaml
  apiRules:
      - name: {FUNCTION_NAME}
        service:
          host: {FUNCTION_NAME}.{DOMAIN_NAME}
        rules:
          - methods:
              - GET
              - POST
              - PUT
              - DELETE
            accessStrategies:
              - handler: noop
  ```

4. Apply the new configuration to the cluster:

  ```bash
  kyma apply function
  ```

5. Check if the Function's code was pushed to the cluster and reflects the local configuration:

  ```bash
  kubectl get apirules $NAME -n $NAMESPACE
  ```

6. Check that the APIRule was created successfully and has the status `OK`:

  ```bash
  kubectl get apirules $NAME -n $NAMESPACE -o=jsonpath='{.status.APIRuleStatus.code}'
  ```

7. Call the Function's external address:

  ```bash
  curl https://$NAME.$DOMAIN
  ```

  </details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Export these variables:

    ```bash
    export DOMAIN={DOMAIN_NAME}
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    ```

    >**NOTE:** Function takes the name from the Function CR name. The APIRule CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

2. Create an APIRule CR for your Function. It is exposed on port `80` that is the default port of the [Service Placeholder](../../05-technical-reference/00-architecture/svls-01-architecture.md).

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      gateway: kyma-system/kyma-gateway
      host: $NAME.$DOMAIN
      rules:
      - path: /.*
        accessStrategies:
        - config: {}
          handler: noop
        methods:
        - GET
        - POST
        - PUT
        - DELETE
      service:
        name: $NAME
        port: 80
    EOF
    ```

3. Check that the APIRule was created successfully and has the status `OK`:

    ```bash
    kubectl get apirules $NAME -n $NAMESPACE -o=jsonpath='{.status.APIRuleStatus.code}'
    ```

4. Access the Function's external address:

    ```bash
    curl https://$NAME.$DOMAIN
    ```

    </details>
    <details>
    <summary label="busola-ui">
    Kyma Dashboard
    </summary>

>**NOTE:** Kyma Dashboard uses Busola, which is not installed by default. Follow the [instructions](https://github.com/kyma-project/busola#installation) to install it.

1. Select a Namespace from the drop-down list in the top navigation panel. Make sure the Namespace includes the Function that you want to expose through an APIRule.

2. Go to **Discovery and Network** > **API Rules**, and click on **Create API Rule**.

3. Enter the following information:

    - The APIRule's **Name** matching the Function's name.

    >**NOTE:** The APIRule CR can have a name different from that of the Function, but it is recommended that all related resources share a common name.

    - **Service Name** matching the Function's name.

    - **Host** to determine the host on which you want to expose your Function. It is required to change the `*` symbol at the beggining to the subdomain name you want.

5. In the **Rules > Access Strategies > Config**  section, change the handler from `allow` to `noop` and select all the methods below.

6. Select **Create** to confirm your changes.

7. Check if you can access the Function by selecting the HTTPS link under the **Host** column for the newly created APIRule.

    </details>
</div>
