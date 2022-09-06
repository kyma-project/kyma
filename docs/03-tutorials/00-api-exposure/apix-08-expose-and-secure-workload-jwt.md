---
title: Expose and secure a workload with JWT
---

This tutorial shows how to expose and secure services or Functions using API Gateway Controller. The Controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio VirtualService and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured workloads, the tutorial uses a JWT token.

You can use it as a follow-up to the [Set up a custom domain for a workload](./apix-02-setup-custom-domain-for-workload.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create them, follow the [Create a workload](./apix-01-create-workload.md) tutorial.
To obtain JWT take a look at [Get a JWT](./apix-06-get-jwt.md) tutorial.

## Expose, secure, and access your workload

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
   >**NOTE:** In this step, you provide `DOMAIN_NAME` which must be a Kyma domain or your custom subdomain, for example, api.mydomain.com. If you don't want to use your custom domain, replace `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`

2. Expose the service and secure it by creating an APIRule CR in your Namespace. Run:

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
       port: 8000
     gateway: $GATEWAY
     rules:
       - accessStrategies:
         - handler: jwt
           config:
             jwks_urls:
             - $JWKS_URI
         methods:
           - GET
         path: /.*
   EOF
   ```

   >**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

3. To access the secured service, call it using the JWT access token:

   ```bash
   curl -ik https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers -H "Authorization: Bearer $ACCESS_TOKEN"
   ```

   This call returns the code `200` response.
   
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
   >**NOTE:** In this step, you provide `DOMAIN_NAME` which must be a Kyma domain or your custom subdomain, for example, api.mydomain.com. If you don't want to use your custom domain, replace `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`

2. Expose the Function and secure it by creating an APIRule CR in your Namespace. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: function
     namespace: $NAMESPACE
   spec:
     host: function-example.$DOMAIN_TO_EXPOSE_WORKLOADS   
     service:
       name: function
       port: 80
     gateway: $GATEWAY
     rules:
       - accessStrategies:
         - handler: jwt
           config:
             jwks_urls:
             - $JWKS_URI
         methods:
           - GET
         path: /.*
   EOF
   ```

3. To access the secured Function, call it using the JWT access token:

   ```bash
   curl -ik https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function -H "Authorization: Bearer $ACCESS_TOKEN"
   ```

   This call returns the code `200` response.

  </details>
</div>
