---
title: Expose a service
---

This tutorial shows how to expose service endpoints and configure different allowed HTTP methods for them using API Gateway Controller.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Deploy a service](./apix-02-deploy-service.md) tutorial.

## Expose and access the resources

Follow the instruction to expose and access your unsecured instance of the HttpBin service or unsecured sample Function.

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Expose the instance of the HttpBin service by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: httpbin
    namespace: $NAMESPACE
  spec:
    service:
      host: httpbin.$DOMAIN
      name: httpbin
      port: 8000
    gateway: namespace-name/httpbin-gateway #The value corresponds to the Gateway CR you created.
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

2. Call the endpoint by sending a `GET` request to the HttpBin service:

  ```bash
  curl -ik -X GET https://httpbin.$DOMAIN/ip
  ```

3. Send a `POST` request to the HttpBin's `/post` endpoint:

  ```bash
  curl -ik -X POST https://httpbin.$DOMAIN/post -d "test data"
  ```

These calls return the code `200` response.

  </details>

  <details>
  <summary>
  Function
  </summary>

1. Expose the sample Function by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

  ```shell
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: function
    namespace: $NAMESPACE
  spec:
    gateway: namespace-name/httpbin-gateway #The value corresponds to the Gateway CR you created. 
    service:
      name: function
      port: 80
      host: function-example.$DOMAIN
    rules:
      - path: /function
        methods: ["GET"]
        accessStrategies:
          - handler: noop
  EOF
  ```

>**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

2. Send a `GET` request to the Function:

  ```shell
  curl -ik https://function-example.$DOMAIN/function
  ```

This call returns the code `200` response.

  </details>
</div>
