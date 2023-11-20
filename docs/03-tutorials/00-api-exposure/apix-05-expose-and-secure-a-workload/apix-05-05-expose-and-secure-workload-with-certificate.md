---
title: Expose and secure a workload with a certificate
---

This tutorial shows how to expose and secure a workload with mutual authentication using TLS Gateway.

## Prerequisites

* Deploy [a sample HttpBin service and sample Function](../apix-01-create-workload.md).
* Set up [your custom domain](../apix-02-setup-custom-domain-for-workload.md).
* Set up [a mutual TLS Gateway](../apix-03-set-up-tls-gateway.md) and export the bundle certificates.
* To learn how to create your own self-signed Client Root CA and Certificate, see [this tutorial](../../00-security/sec-02-mtls-selfsign-client-certicate.md). This step is optional.

## Authorize a client with a certificate

The following instructions describe how to secure an mTLS service or a Function. 
>**NOTE:** Create AuthorizationPolicy to check if the client's common name in the certificate matches.

1. Export the following values as environment variables:

   ```bash
   export CLIENT_ROOT_CA_CRT_FILE={CLIENT_ROOT_CA_CRT_FILE}
   export CLIENT_CERT_CN={COMMON_NAME}
   export CLIENT_CERT_ORG={ORGANIZATION}
   export CLIENT_CERT_CRT_FILE={CLIENT_CERT_CRT_FILE}
   export CLIENT_CERT_KEY_FILE={CLIENT_CERT_KEY_FILE}
   ```
2. Create VirtualService that adds the X-CLIENT-SSL headers to incoming requests:

<!-- tabs:start -->

#### **HttpBin**

Run:

   ```bash
   cat <<EOF | kubectl apply -f - 
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: httpbin-vs
     namespace: ${NAMESPACE}
   spec:
     hosts:
     - "httpbin-vs.${DOMAIN_TO_EXPOSE_WORKLOADS}"
     gateways:
     - ${MTLS_GATEWAY_NAME}
     http:
     - route:
       - destination:
           port:
             number: 8000
           host: httpbin
         headers:
           request:
             set:
               X-CLIENT-SSL-CN: "%DOWNSTREAM_PEER_SUBJECT%"
               X-CLIENT-SSL-SAN: "%DOWNSTREAM_PEER_URI_SAN%"
               X-CLIENT-SSL-ISSUER: "%DOWNSTREAM_PEER_ISSUER%"
   EOF
   ```

#### **Function**

Run:
   ```bash
   cat <<EOF | kubectl apply -f - 
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: function-vs
     namespace: ${NAMESPACE}
   spec:
     hosts:
     - "function-vs.${DOMAIN_TO_EXPOSE_WORKLOADS}"
     gateways:
     - ${MTLS_GATEWAY_NAME}
     http:
     - route:
       - destination:
           port:
             number: 80
           host: function
         headers:
           request:
             set:
               X-CLIENT-SSL-CN: "%DOWNSTREAM_PEER_SUBJECT%"
               X-CLIENT-SSL-SAN: "%DOWNSTREAM_PEER_URI_SAN%"
               X-CLIENT-SSL-ISSUER: "%DOWNSTREAM_PEER_ISSUER%"
   EOF
   ```

<!-- tabs:end -->

1. Create AuthorizationPolicy that verifies if the request contains a client certificate:

<!-- tabs:start -->

#### **HttpBin**

Run:
    
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: security.istio.io/v1beta1
   kind: AuthorizationPolicy
   metadata:
     name: test-authz-policy
     namespace: ${NAMESPACE}
   spec:
     action: ALLOW
     rules:
     - to:
       - operation:
           hosts: ["httpbin-vs.${DOMAIN_TO_EXPOSE_WORKLOADS}"]
       when:
       - key: request.headers[X-Client-Ssl-Cn]
         values: ["O=${CLIENT_CERT_ORG},CN=${CLIENT_CERT_CN}"]
   EOF
   ```

#### **Function**

Run:
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: security.istio.io/v1beta1
   kind: AuthorizationPolicy
   metadata:
     name: test-authz-policy
     namespace: ${NAMESPACE}
   spec:
     action: ALLOW
     rules:
     - to:
       - operation:
           hosts: ["function-vs.${DOMAIN_TO_EXPOSE_WORKLOADS}"]
       when:
       - key: request.headers[X-Client-Ssl-Cn]
         values: ["O=${CLIENT_CERT_ORG},CN=${CLIENT_CERT_CN}"]
   EOF
   ```

<!-- tabs:end -->

1. Call the secured endpoints of the HttpBin service or the secured Function.

<!-- tabs:start -->

#### **HttpBin**

  Send a `GET` request to the HttpBin service with the client certificates that you used to create mTLS Gateway:

   ```shell
   curl --key ${CLIENT_CERT_KEY_FILE} \
         --cert ${CLIENT_CERT_CRT_FILE} \
         --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
         -ik -X GET https://httpbin-vs.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
   ```

  If successful, the call returns the code `200 OK` response. If you call the service without the proper certificates or with invalid ones, you get the code `403` response.

#### **Function**

  Send a `GET` request to the Function with the client certificates that you used to create mTLS Gateway:

   ```shell
   curl --key ${CLIENT_CERT_KEY_FILE} \
         --cert ${CLIENT_CERT_CRT_FILE} \
         --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
         -ik -X GET https://function-vs.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```

  If successful, the call returns the code `200 OK` response. If you call the Function without the proper certificates or with invalid ones, you get the code `403` response.

<!-- tabs:end -->