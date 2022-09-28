---
title: Expose a workload
---

This tutorial shows how to expose service endpoints and configure different allowed HTTP methods for them using API Gateway Controller.
   > **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [OAuth2](./apix-05-expose-and-secure-workload-oauth2.md) or [JWT](./apix-08-expose-and-secure-workload-jwt.md).

The tutorial may be a follow-up to the [Set up a custom domain for a workload](./apix-02-setup-custom-domain-for-workload.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

## Expose and access your workload

Follow the instruction to expose and access your unsecured instance of the HttpBin service or unsecured sample Function.

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/httpbin-gateway
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace `DOMAIN_NAME` with a Kyma domain and `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`

2. Expose the instance of the HttpBin service by creating an APIRule CR in your Namespace. Run:

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

4. Send a `POST` request to the HttpBin's `/post` endpoint:

   ```bash
   curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data"
   ```

   These calls return the code `200` response.

  </details>

  <details>
  <summary>
  Function
  </summary>

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/httpbin-gateway 
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace `DOMAIN_NAME` with a Kyma domain and `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`

2. Expose the sample Function by creating an APIRule CR in your Namespace. Run:

   ```shell
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

   ```shell
   curl -ik https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```

   This call returns the code `200` response.

  </details>
</div>
