---
title: Expose and secure a workload with a certificate 
---

This tutorial shows how to expose a workload with mutual authentication using [`kyma-mtls-gateway`](https://github.com/kyma-project/kyma/blob/main/resources/certificates/templates/mtls-certificate.yaml).

The tutorial may be a follow-up to the [Set up a custom domain for a workload](./apix-02-setup-custom-domain-for-workload.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

Before you start, you must set up [`kyma-mtls-gateway`](../00-security/sec-02-setup-mtls-gateway.md) to allow mutual authentication in Kyma. 

## Expose and access your workload

Follow the instruction to expose and access your instance of the HttpBin service or your sample Function.
  > **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [OAuth2](./apix-05-expose-and-secure-workload-oauth2.md) or [JWT](./apix-08-expose-and-secure-workload-jwt.md).
<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace `DOMAIN_NAME` with a Kyma domain

2. Expose the instance of the HttpBin service by creating an APIRule CR in your Namespace. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     host: httpbin.mtls.$DOMAIN_TO_EXPOSE_WORKLOADS
     service:
       name: httpbin
       port: 8000
     gateway: kyma-system/kyma-mtls-gateway
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
   >**NOTE:** If you are running Kyma on k3d, add `httpbin.mtls.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

3. Call the endpoint by sending a `GET` request to the HttpBin service:

   ```bash
   curl --key client.key \
      --cert client.crt 
      --cacert client-root-ca.crt \
      -ik -X GET https://httpbin.mtls.$DOMAIN_TO_EXPOSE_WORKLOADS/ip
   ```
   
4. Send a `POST` request to the HttpBin's `/post` endpoint:

   ```bash
   curl --key client.key \
      --cert client.crt 
      --cacert client-root-ca.crt \
      -ik -X POST https://httpbin.mtls.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data"
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
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace `DOMAIN_NAME` with a Kyma domain

2. Expose the sample Function by creating an APIRule CR in your Namespace. Run:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: function
     namespace: $NAMESPACE
   spec:
     gateway: kyma-system/kyma-mtls-gateway
     host: function-example.mtls.$DOMAIN_TO_EXPOSE_WORKLOADS
     service:
       name: function
       port: 80
     rules:
       - path: /function
         methods: ["GET"]
         accessStrategies:
           - handler: noop
   EOF
   ```

3. Send a `GET` request to the Function:

   ```shell
   curl --key client.key \
      --cert client.crt 
      --cacert client-root-ca.crt \
      -ik -X GET https://function-example.mtls.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```
   This call returns the code `200` response.
  </details>
</div>
