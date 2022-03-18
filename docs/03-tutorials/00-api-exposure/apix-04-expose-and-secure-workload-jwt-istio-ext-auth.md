---
title: Expose and secure a workload with Istio external auth
---

This tutorial shows how to expose workload using Istio authorization policy and secure it using Istio RequestAuthentication with JWT token

## Prerequisites

To follow this tutorial, use Kyma 2.0 or higher.

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-02-create-workload.md) tutorial.

## Add a RequestAuthentication which requires JWT token for all request for workloads that have label app:httpbin

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


## Add a different JWT requirement for a different host and fine tune the authorization policy to set different requirement per path

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
     - issuer: "issuer-bar"
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
           paths: ["/healthz"]
   EOF
   ```
