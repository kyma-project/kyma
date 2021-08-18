---
title: Expose a Function with an APIRule
---

This tutorial shows how you can expose your Function to access it outside the cluster, through an HTTP proxy. To expose it, use an APIRule custom resource (CR) managed by the in-house API Gateway Controller. This controller reacts to an instance of the APIRule CR and, based on its details, it creates an Istio Virtual Service and Oathkeeper Access Rules that specify your permissions for the exposed Function.

When you complete this tutorial, you get a Function that:

- Is available on an unsecured endpoint (**handler** set to `noop` in the APIRule CR).
- Accepts the `GET`, `POST`, `PUT`, and `DELETE` methods.

>**NOTE:** To learn more about securing your Function, see the [tutorial](../api-exposure/apig-01-expose-and-secure-service.md).

>**TIP:** Read also about [Function’s specification](../../05-technical-reference/svls-08-function-specification.md) if you are interested in its signature, `event` and `context` objects, and custom HTTP responses the Function returns.

## Prerequisites

This tutorial is based on an existing Function. To create one, follow the [Create a Function](./svls-01-create-inline-function.md) tutorial.

## Steps

Follows these steps:

<div tabs name="steps" group="expose-function">
  <details>
  <summary label="cli">
  Kyma CLI
  </summary>

1. Export these variables:

      ```bash
      export DOMAIN={DOMAIN_NAME}
      export NAME={APIRULE_NAME}
      ```

2. Download the latest configuration of the Function from the cluster. This way you will update the local `config.yaml` file with the Function's code.

  ```bash
  kyma sync function $NAME -n $NAMESPACE
  ```

3. Edit the local `config.yaml` file and add the **apiRules** schema for the Function at the end of the file:

  ```yaml
  apiRules:
      - name: {APIRULE_NAME}
        service:
          host: {APIRULE_NAME}.{DOMAIN_NAME}
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

2. Create an APIRule CR for your Function. It is exposed on port `80` that is the default port of the [Service Placeholder](../../05-technical-reference/03-architecture/svls-01-architecture.md).

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1alpha1
    kind: APIRule
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      gateway: kyma-gateway.kyma-system.svc.cluster.local
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

3. Check that the APIRule was created successfully and has the status `OK`:

    ```bash
    kubectl get apirules $NAME -n $NAMESPACE -o=jsonpath='{.status.APIRuleStatus.code}'
    ```

4. Access the Function's external address:

    ```bash
    curl https://$NAME.$DOMAIN
    ```

   >**CAUTION:** If you have a Minikube cluster, you must first add its IP address mapped to the hostname of the exposed Kubernetes Service to the `hosts` file on your machine.

    </details>
    <details>
    <summary label="busola-ui">
    Busola UI
    </summary>

>**NOTE:** Busola is not installed by default. Follow the [instructions](https://github.com/kyma-project/busola#installation) to install it.

1. Select a Namespace from the drop-down list in the top navigation panel. Make sure the Namespace includes the Function that you want to expose through an APIRule.

2. In the left navigation panel, go to **Workloads** > **Functions** and select the Function you want to expose.

3. Switch to the **Configuration** tab and select **Expose Function** in the APIRules section. A pop-up box with the form will appear on the screen.

4. In the **General settings** section of the pop-up box:

    - Enter the APIRule's **Name** matching the Function's name.

    >**NOTE:** The APIRule CR can have a name different from that of the Function, but it is recommended that all related resources share a common name.

    - Enter **Hostname** to indicate the host on which you want to expose your Function.

5. In the **Access strategies** section, select the `noop` handler from the drop-down list and leave the default settings with the `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, and `HEAD` methods.

6. Select **Create** to confirm changes. The pop-up box with the form will close.

7. Check if you can access the Function by selecting the HTTPS link under the **Host** column for the newly created APIRule.

    >**CAUTION:** If you have a Minikube cluster, you must first add its IP address mapped to the hostname of the exposed Kubernetes Service to the `hosts` file on your machine.

    </details>
</div>
