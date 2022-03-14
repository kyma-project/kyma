---
title: Expose a Function with an API Rule
---

This tutorial shows how you can expose your Function to access it outside the cluster, through an HTTP proxy. To expose it, use an API Rule custom resource (CR) managed by the in-house API Gateway Controller. This controller reacts to an instance of the API Rule CR and, based on its details, it creates an Istio Virtual Service and Oathkeeper Access Rules that specify your permissions for the exposed Function.

When you complete this tutorial, you get a Function that:

- Is available on an unsecured endpoint (**handler** set to `noop` in the API Rule CR).
- Accepts the `GET`, `POST`, `PUT`, and `DELETE` methods.

>**NOTE:** To learn more about securing your Function, see the [Expose and secure a workload with OAuth2](../00-api-exposure/apix-03-expose-and-secure-workload-oauth2.md) tutorial.

>**TIP:** Read also about [Function’s specification](../../05-technical-reference/svls-08-function-specification.md) if you are interested in its signature, `event` and `context` objects, and custom HTTP responses the Function returns.

## Prerequisites

This tutorial is based on an existing Function. To create one, follow the [Create a Function](./svls-01-create-inline-function.md) tutorial.

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
   >**NOTE:** The Function takes the name from the Function CR name. The API Rule CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.
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

6. Check that the API Rule was created successfully and has the status `OK`:

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

    >**NOTE:** Function takes the name from the Function CR name. The API Rule CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

2. Create an API Rule CR for your Function. It is exposed on port `80` that is the default port of the [Service Placeholder](../../05-technical-reference/00-architecture/svls-01-architecture.md).

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1alpha1
    kind: APIRule
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      gateway: kyma-system/kyma-gateway
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
        host: $NAME.$DOMAIN
        name: $NAME
        port: 80
    EOF
    ```

3. Check that the API Rule was created successfully and has the status `OK`:

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

1. Select a Namespace from the drop-down list in the top navigation panel. Make sure the Namespace includes the Function that you want to expose through an API Rule.

2. Go to **Workloads** > **Functions** and select the Function you want to expose.

3. Switch to the **Configuration** tab and select **Create API Rule** in the API Rules section.

4. Under **General settings**, enter the following information:

    - The API Rule's **Name** matching the Function's name.

    >**NOTE:** The API Rule CR can have a name different from that of the Function, but it is recommended that all related resources share a common name.

    - **Subdomain** to determine the host on which you want to expose your Function.

5. In the **Rules** section, select the `noop` handler and mark **all** the methods.

6. Select **Create** to confirm your changes.

7. Check if you can access the Function by selecting the HTTPS link under the **Host** column for the newly created API Rule.

    </details>
</div>
