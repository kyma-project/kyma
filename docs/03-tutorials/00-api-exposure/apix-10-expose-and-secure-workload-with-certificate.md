---
title: Expose and secure a workload with a certificate
---

This tutorial shows how to expose and secure a workload with mutual authentication using a mutual TLS Gateway.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

Before you start, Set up:
- [Custom Domain](./apix-02-setup-custom-domain-for-workload.md) - skip step 5 (Create a Gateway CR)
- [mTLS Gateway](../00-security/sec-03-setup-mtls-gateway.md) to allow mutual authentication in Kyma and make sure that you exported the [bundle certificates](../00-security/sec-03-setup-mtls-gateway#steps).

Optionally, take a look at the [How to create own self-signed Client Root CA and Certificate](../00-security/sec-02-mtls-selfsign-client-certicate.md) tutorial.

## Authorize client with a certificate

The following instructions describe how to further secure the mTLS service or Function. 
>**NOTE:** Create AuthorizationPolicy to check if the client's common name in the certificate matches.

1. Export the following values as environment variables:

   ```bash
   export CLIENT_ROOT_CA_CRT_FILE={CLIENT_ROOT_CA_CRT_FILE}
   export CLIENT_CERT_CN={COMMON_NAME}
   export CLIENT_CERT_ORG={ORGANIZATION}
   export CLIENT_CERT_CRT_FILE={CLIENT_CERT_CRT_FILE}
   export CLIENT_CERT_KEY_FILE={CLIENT_CERT_KEY_FILE}
   ```

<div tabs>
  <details>
  <summary>
  HttpBin
  </summary>

1. Create VirtualService that adds the X-CLIENT-SSL headers to the incoming requests:
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

2. Create AuthorizationPolicy that verifies if the request contains a client certificate:
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
  </details>
  <details>
  <summary>
  Function
  </summary>

1. Create VirtualService that adds the X-CLIENT-SSL headers to incoming requests:
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
2. Create AuthorizationPolicy that verifies if the request contains a client certificate:
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
  </details>
</div>

<div tabs>

  <details>
  <summary>
  Call the secured endpoints of a service
  </summary>

Send a `GET` request to the HttpBin service with the client certificates that you used to create mTLS Gateway:

   ```shell
   curl --key ${CLIENT_CERT_KEY_FILE} \
        --cert ${CLIENT_CERT_CRT_FILE} \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
        -ik -X GET https://httpbin-vs.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
   ```

These calls return the code `200` response. If you call the service without the proper certificates or with invalid ones, you get the code `403` response.

  </details>

  <details>
  <summary>
  Call the secured Function
  </summary>

Send a `GET` request to the Function with the client certificates that you used to create mTLS Gateway:

   ```shell
   curl --key ${CLIENT_CERT_KEY_FILE} \
        --cert ${CLIENT_CERT_CRT_FILE} \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
        -ik -X GET https://function-vs.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```

This call returns the code `200` response. If you call the Function without the proper certificates or with invalid ones, you get the code `403` response.
  </details>
</div>