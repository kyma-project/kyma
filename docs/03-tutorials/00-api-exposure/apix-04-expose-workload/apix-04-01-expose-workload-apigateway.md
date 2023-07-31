---
title: Expose a workload
---

This tutorial shows how to expose an unsecured instance of the HttpBin service or an unsecured sample Function and call their endpoints.

   > **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [OAuth2](../apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md) or [JWT](../apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md)).

## Prerequisites

* Deploy [a sample HttpBin service and a sample Function](../apix-01-create-workload.md).
* Set up [your custom domain](../apix-02-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 
* Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
  <div tabs name="export-values">

    <details>
    <summary>
    Custom domain
    </summary>
    
    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export GATEWAY=$NAMESPACE/httpbin-gateway
    ```
    </details>

    <details>
    <summary>
    Kyma domain
    </summary>

    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS=local.kyma.dev
    export GATEWAY=kyma-system/kyma-gateway
    ```
    </details>
  </div>   
   
## Expose and access your workload

Follow these steps:

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Expose an instance of the HttpBin service by creating APIRule CR in your Namespace. Run:

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

2. Call the endpoint by sending a `GET` request to the HttpBin service:

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/ip
    ```
  If successful, the call returns the code `200 OK` response.

3. Call the endpoint by sending a `POST` request to the HttpBin service:

    ```bash
    curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data"
    ```
  If successful, the call returns the code `200 OK` response.

  </details>

  <details>
  <summary>
  Function
  </summary>

1. Expose the sample Function by creating APIRule CR in your Namespace. Run:

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

2. Send a `GET` request to the Function:

    ```bash
    curl -ik https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function
    ```

  If successful, the call returns the code `200 OK` response.

  </details>
</div>
