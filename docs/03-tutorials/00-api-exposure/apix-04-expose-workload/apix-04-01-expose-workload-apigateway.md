---
title: Expose a workload
---

This tutorial shows how to expose an unsecured instance of the HttpBin service or an unsecured sample Function and call their endpoints.

   > **CAUTION:** Exposing a workload to the outside world causes a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [OAuth2](../00-api-exposure/apix-07-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md) or [JWT](../00-api-exposure/apix-07-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md).

## Prerequisites

* [Sample HttpBin service and sample Function](./apix-01-create-workload.md) deployed
* If you want to use your custom domain instead of a Kyma domain, follow [this tutorial](./apix-02-setup-custom-domain-for-workload.md) to learn how to set it up.

## Expose and access your workload

Follow these steps:

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Export the following values as environment variables.

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/httpbin-gateway
   ```
   >**NOTE:** The `DOMAIN_NAME` refers to the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace the `DOMAIN_NAME` with a Kyma domain and the `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`.

2. Expose an instance of the HttpBin service by creating APIRule CR in your Namespace. Run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: httpbin
      namespace: $NAMESPACE
    spec:
      host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
      service:
        name: httpbin
        namespace: $NAMESPACE
        port: 8000
      gateway: $GATEWAY
      rules:
        - path: /.*
          methods: ["GET"]
          accessStrategies:
            - handler: noop
          mutators:
            - handler: noop
        - path: /post
          methods: ["POST"]
          accessStrategies:
            - handler: noop
          mutators:
            - handler: noop
    EOF
    ```
  
   >**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

   >**NOTE:** If you don't specify a Namespace for your service, the default APIRule Namespace is used.

3. Call the endpoint by sending a `GET` request to the HttpBin service:

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/ip
    ```
  The call should return the code `200` response.

4. Call the endpoint by sending a `POST` request to the HttpBin service:

    ```bash
    curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data"
    ```
  The call should return the code `200` response.

  </details>

  <details>
  <summary>
  Function
  </summary>

1. Export the following values as environment variables:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/httpbin-gateway 
   ```
   >**NOTE:** The `DOMAIN_NAME` refers to the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace the `DOMAIN_NAME` with a Kyma domain and the `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`.

2. Expose the sample Function by creating APIRule CR in your Namespace. Run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: function
      namespace: $NAMESPACE
    spec:
      gateway: $GATEWAY
      host: function-example.$DOMAIN_TO_EXPOSE_WORKLOADS
      service:
        name: function
        namespace: $NAMESPACE
        port: 80
      rules:
        - path: /function
          methods: ["GET"]
          accessStrategies:
            - handler: noop
    EOF
    ```

   >**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

   >**NOTE:** If you don't specify a Namespace for your service, the default APIRule Namespace is used.

3. Send a `GET` request to the Function:

    ```bash
    curl -ik https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function
    ```

  The call should return the code `200` response.

  </details>
</div>