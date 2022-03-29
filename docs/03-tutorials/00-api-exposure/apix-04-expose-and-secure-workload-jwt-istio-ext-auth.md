---
title: Expose and secure a workload with Istio external auth
---

This tutorial shows how to expose workload using VirtualService and secure it using Istio authorization policy together with Istio RequestAuthentication based on JWT token

## Prerequisites

To follow this tutorial, use Kyma 2.0 or higher.

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-02-create-workload.md) tutorial.

## Expose your workload using VirtualService

1. Export environment variable with your namespace:

   ```shell
   export NAMESPACE={NAMESPACE} # e.g. default
   export DOMAIN={DOMAIN_NAME} # This is a Kyma domain or your custom subdomain e.g. api.mydomain.com.
   ```

1. Run:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     hosts:
     - "httpbin.$DOMAIN"
     gateways:
     - kyma-system/kyma-gateway # or the Gateway CR you created
     http:
     - match:
       - uri:
           prefix: /
       route:
       - destination:
           port:
             number: 8000
           host: httpbin.default.svc.cluster.local
   EOF
   ```

## Add a RequestAuthentication which requires JWT token for all requests for workloads that have label app:httpbin

1. Export the following values:

   ```shell
   export JWKSURI={YOURJWKSURL} # e.g. https://example.com/.well-known/jwks.json
   ```

1. Run:

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
    - issuer: issuer
      jwksUri: $JWKSURI
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

2. If you try to access secured workload you should get 403 Forbidden error:

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN/status/200
   ```

3. Using correct JWT token should give you 200 OK response

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN/status/200 --header 'Authorization:Bearer JWT_TOKEN'
   ```

## Add a different JWT requirement for a different host

1. Run:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: security.istio.io/v1beta1
   kind: RequestAuthentication
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     selector:
       matchLabels:
         app: httpbin
     jwtRules:
     - issuer: "issuer-foo"
       jwksUri: https://example.com/.well-known/jwks.json
     - issuer: "issuer-bar"
       jwksUri: https://example.com/.well-known/jwks.json
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
           requestPrincipals: ["issuer-foo/*"]
       to:
       - operation:
           hosts: ["example.com"]
     - from:
       - source:
           requestPrincipals: ["issuer-bar/*"]
       to:
       - operation:
           hosts: ["another-host.com"]
   EOF
   ```
