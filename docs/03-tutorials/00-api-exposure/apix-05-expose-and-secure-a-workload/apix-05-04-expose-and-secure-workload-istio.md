---
title: Expose and secure a workload with Istio
---

This tutorial shows how to expose and secure a workload using Istio's built-in security features. You will expose the workload by creating a [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/). Then, you will secure access to your workload by adding the JWT validation verified by the Istio security configuration with [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/).

## Prerequisites

* [Sample HttpBin service and sample Function](../apix-01-create-workload.md) deployed
* [JSON Web Token (JWT)](./apix-05-02-get-jwt.md).
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

## Expose your workload using a Virtual Service

Follow the instructions in the tabs to expose the HttpBin workload or the Function using a VirtualService.

<!-- tabs:start -->

#### **Expose the HttpBin workload**

1. Create a VirtualService:

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

#### **Expose the Function**

1. Create a VirtualService:

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

<!-- tabs:end -->

## Secure a workload or the Function using a JWT

To secure the HttpBin workload or the Function using a JWT, create a Request Authentication with Authorization Policy. Workloads with the `matchLabels` parameter specified require a JWT for all requests. Follow the instructions in the tabs:

<!-- tabs:start -->

#### **Secure the HttpBin workload**

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

2. Access the workload you secured. You get the code `403 Forbidden` error.

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200
   ```

3. Now, access the secured workload using the correct JWT. You get the code `200 OK` response.

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200 --header "Authorization:Bearer $ACCESS_TOKEN"
   ```

#### **Secure the Function**

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

2. Access the workload you secured. You get the code `403 Forbidden` error.

   ```shell
   curl -ik -X GET https://function.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200
   ```

3. Now, access the secured workload using the correct JWT. You get the code `200 OK` response.

   ```shell
   curl -ik -X GET https://function.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200 --header "Authorization:Bearer $ACCESS_TOKEN"
   ```

<!-- tabs:end -->