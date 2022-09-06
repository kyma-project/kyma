---
title: Expose and secure a workload with Istio
---

This tutorial shows how to expose and secure a workload using Istio built-in security features. You will expose the workload by creating a [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/). Then, you will secure access to your workload by adding the JWT validation verified by the Istio security configuration with [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/).

## Prerequisites

- You have got a JSON Web Token (JWT). For more information, see [Get a JWT](./apix-06-get-jwt.md).

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-01-create-workload.md) tutorial. It can also be a follow-up to the [Set up a custom domain for a workload](./apix-02-setup-custom-domain-for-workload.md) tutorial.

## Expose your workload using a Virtual Service

Follow the instructions in the tabs to expose the HttpBin workload or the Function using a VirtualService.

<div tabs>

  <details>
  <summary>
  Expose the HttpBin workload
  </summary>

1. Export your domain name and gateway name as environment variables:

   ```shell
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/httpbin-gateway 
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace `DOMAIN_NAME` with a Kyma domain and `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`

2. Create a VirtualService:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     hosts:
     - "httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS"
     gateways:
     - $GATEWAY
     http:
     - match:
       - uri:
           prefix: /
       route:
       - destination:
           port:
             number: 8000
           host: httpbin.$NAMESPACE.svc.cluster.local
   EOF
   ```
  </details>

  <details>
  <summary>
  Expose the Function
  </summary>

1. Export your domain name and gateway name as environment variables:

   ```shell
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/httpbin-gateway 
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace `DOMAIN_NAME` with a Kyma domain and `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`

2. Create a VirtualService:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: function
     namespace: $NAMESPACE
   spec:
     hosts:
     - "function.$DOMAIN_TO_EXPOSE_WORKLOADS"
     gateways:
     - $GATEWAY
     http:
     - match:
       - uri:
           prefix: /
       route:
       - destination:
           port:
             number: 80
           host: function.$NAMESPACE.svc.cluster.local
   EOF
   ```

  </details>
</div>

## Secure a workload or the Function using a JWT

To secure the Httpbin workload or the Function using a JWT, create a Request Authentication with Authorization Policy. Workloads that have the **matchLabels** parameter specified require a JWT for all requests. Follow the instructions in the tabs:

<div tabs>

  <details>
  <summary>
  Secure the Httpbin workload
  </summary>

1. Create the Request Authentication and Authorization Policy resources:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: security.istio.io/v1beta1
   kind: RequestAuthentication
   metadata:
     name: jwt-auth-httpbin
     namespace: $NAMESPACE
   spec:
     selector:
       matchLabels:
         app: httpbin
     jwtRules:
     - issuer: $ISSUER
       jwksUri: $JWKS_URI
   ---
   apiVersion: security.istio.io/v1beta1
   kind: AuthorizationPolicy
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     selector:
       matchLabels:
         app: httpbin
     rules:
     - from:
       - source:
           requestPrincipals: ["*"]
   EOF
   ```

2. Access the workload you secured. You will get the `403 Forbidden` error.

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200
   ```

3. Now, access the secured workload using the correct JWT. You will get the `200 OK` response.

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200 --header "Authorization:Bearer $ACCESS_TOKEN"
   ```
  </details>

  <details>
  <summary>
  Secure the Function
  </summary>

1. Create the Request Authentication and Authorization Policy resources:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: security.istio.io/v1beta1
   kind: RequestAuthentication
   metadata:
     name: jwt-auth-function
     namespace: $NAMESPACE
   spec:
     selector:
       matchLabels:
         app: function
     jwtRules:
     - issuer: $ISSUER
       jwksUri: $JWKS_URI
   ---
   apiVersion: security.istio.io/v1beta1
   kind: AuthorizationPolicy
   metadata:
     name: function
     namespace: $NAMESPACE
   spec:
     selector:
       matchLabels:
         app: function
     rules:
     - from:
       - source:
           requestPrincipals: ["*"]
   EOF
   ```

2. Access the workload you secured. You will get the `403 Forbidden` error.

   ```shell
   curl -ik -X GET https://function.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200
   ```

3. Now, access the secured workload using the correct JWT. You will get the `200 OK` response.

   ```shell
   curl -ik -X GET https://function.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200 --header "Authorization:Bearer $ACCESS_TOKEN"
   ```
  </details>
</div>
