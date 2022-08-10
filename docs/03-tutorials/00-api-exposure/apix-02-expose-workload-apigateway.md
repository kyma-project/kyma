---
title: Expose a workload
---

This tutorial shows how to expose service endpoints and configure different allowed HTTP methods for them using API Gateway Controller.

The tutorial may be a follow-up to the [Use a custom domain to expose a workload](./apix-01-own-domain.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-02-create-workload.md) tutorial.

## Expose and access your workload

Follow the instruction to expose and access your unsecured instance of the HttpBin service or unsecured sample Function.

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} #This is a Kyma domain or your custom subdomain e.g. api.mydomain.com.
   ```

2. Expose the instance of the HttpBin service by creating an APIRule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1alpha1
   kind: APIRule
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     service:
       host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
       name: httpbin
       port: 8000
     gateway: $NAMESPACE/httpbin-gateway #The value corresponds to the Gateway CR you created.
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
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} #This is a Kyma domain or your custom subdomain e.g. api.mydomain.com.
   ```

2. Expose the sample Function by creating an APIRule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

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
       host: function-example.$DOMAIN_TO_EXPOSE_WORKLOADS
     rules:
       - path: /function
         methods: ["GET"]
         accessStrategies:
           - handler: noop
   EOF
   ```

   >**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

3. Send a `GET` request to the Function:

   ```shell
   curl -ik https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```

   This call returns the code `200` response.

  </details>
</div>
