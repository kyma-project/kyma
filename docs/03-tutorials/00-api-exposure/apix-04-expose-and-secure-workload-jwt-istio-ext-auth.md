---
title: Expose and secure a workload with Istio external auth
---

This tutorial shows how to expose workload using Istio authorization policy and secure it using Istio RequestAuthentication with JWT token

## Prerequisites

To follow this tutorial, use Kyma 2.0 or higher.

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-02-create-workload.md) tutorial.

## Expose your workload using VirtualService

1. Export environment variable with your namespace:

   ```shell
   export NAMESPACE={NAMESPACE} # e.g. default
   ```

1. Run:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1alpha3
   kind: Gateway
   metadata:
     name: httpbin-gateway
     namespace: istio-system
   spec:
     selector:
       istio: ingressgateway
     servers:
     - port:
         number: 80
         name: http
         protocol: HTTP
       hosts:
       - "*"
   ---
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     hosts:
     - "*"
     gateways:
     - istio-system/httpbin-gateway
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

2. Export the following values:

   ```shell
   export DOMAIN={DOMAIN_NAME} # This is a Kyma domain or your custom subdomain e.g. api.mydomain.com.
   export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
   ```

3. Create DNSEntry with your domain address:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: dns.gardener.cloud/v1alpha1
   kind: DNSEntry
   metadata:
     name: dns-httpbin
     namespace: $NAMESPACE
     annotations:
       dns.gardener.cloud/class: garden
   spec:
     dnsName: "$DOMAIN"
     ttl: 600
     targets:
       - $IP
   EOF
   ```

## Add a RequestAuthentication which requires JWT token for all request for workloads that have label app:httpbin

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
    - issuer: "issuer-foo"
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
           requestPrincipals: ["*"]
   EOF
   ```

2. If you try to access secured workload you should get 403 Forbidden error:

   ```shell
   curl -ik -X GET $DOMAIN/status/200
   ```

3. Using correct JWT token should give you 200 OK response

   ```shell
   curl -ik -X GET $DOMAIN/status/200 --header 'Authorization:Bearer JWT_TOKEN'
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
