---
title: Expose a Function with an API Rule
type: Deep Dive
---

This tutorial shows how you can expose your Function to access it outside the cluster, through an HTTP proxy. To expose it, use an APIRule custom resource (CR) managed by the in-house API Gateway Controller. This controller reacts to an instance of the APIRule CR and, based on its details, it creates an Istio Virtual Service and Oathkeeper Access Rules that specify your permissions for the exposed Function.

When you complete this tutorial, you get a Function that:

- Is available on an unsecured endpoint (**handler** set to `noop` in the APIRule CR).
- Accepts the `GET`, `POST`, `PUT`, and `DELETE` methods.

>**NOTE:** To learn more about securing your Function, see the [tutorial](/components/api-gateway#tutorials-expose-and-secure-a-service-deploy-expose-and-secure-the-sample-resources).

## Prerequisites

This tutorial is based on an existing Function. To create one, follow the [Create a Function](#tutorials-create-a-function) tutorial.

## Steps

Follows these steps:

<div tabs name="steps" group="expose-function">
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

2. Create an APIRule CR for your Function. It is exposed on port `80` that is the default port of the [Service Placeholder](#architecture-architecture).

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

3. Check that the API Rule was created successfully and has the status `OK`:

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
    <summary label="console-ui">
    Console UI
    </summary>

1. Select a Namespace from the drop-down list in the top navigation panel. Make sure the Namespace includes the Function that you want to expose through an API Rule.

2. In the left navigation panel, go to **Workloads** > **Functions** and select the Function you want to expose.

3. Switch to the **Configuration** tab and select **Expose Function** in the **API Rules** section. A pop-up box with the form will appear on the screen.

4. In the **General settings** section:

    - Enter the API Rule's **Name** matching the Function's name.

    >**NOTE:** The APIRule CR can have a name different from that of the Function, but it is recommended that all related resources share a common name.

    - Enter **Hostname** to indicate the host on which you want to expose your Function.

5. In the **Access strategies** section, leave the default settings with the `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, and `HEAD` methods and the `noop` handler selected.

6. Select **Create** to confirm changes. The pop-up box with the form will close.

7. Check if you can access the Function by selecting the HTTPS link under the **Host** column for the newly created API Rule.

    >**CAUTION:** If you have a Minikube cluster, you must first add its IP address mapped to the hostname of the exposed Kubernetes Service to the `hosts` file on your machine.

    </details>
</div>
