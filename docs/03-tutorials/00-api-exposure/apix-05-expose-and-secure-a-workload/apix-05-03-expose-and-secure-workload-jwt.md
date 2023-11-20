---
title: Expose and secure a workload with JWT
---

This tutorial shows how to expose and secure services or Functions using API Gateway Controller. The Controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio VirtualService and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured workloads, the tutorial uses a JWT token.

## Prerequisites

* [Sample HttpBin service and sample Function](../apix-01-create-workload.md) deployed
* [JSON Web Token (JWT)](./apix-05-02-get-jwt.md)
* Set up [your custom domain](../apix-02-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 
* Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:

<!-- tabs:start -->

#### **Custom domain**

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/httpbin-gateway
   ```

#### **Kyma domain**

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
   export GATEWAY=kyma-system/kyma-gateway
   ```

<!-- tabs:end -->

## Expose, secure, and access your workload

<!-- tabs:start -->

#### **HttpBin**

1. Expose the service and secure it by creating an APIRule CR in your Namespace. Run:

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

2. To access the secured service, call it using the JWT access token:

   ```bash
   curl -ik https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers -H "Authorization: Bearer $ACCESS_TOKEN"
   ```

  If successful, the call returns the code `200 OK` response.

#### **Function**

1. Expose the Function and secure it by creating an APIRule CR in your Namespace. Run:

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

  If successful, the call returns the code `200 OK` response.

<!-- tabs:end -->