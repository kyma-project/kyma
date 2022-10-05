---
title: Expose and secure a workload with a certificate
---

This tutorial shows how to expose and secure a workload with mutual authentication using Kyma's mutual TLS Gateway.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

Before you start, Set up [mTLS Gateway](../00-security/sec-02-setup-mtls-gateway.md) to allow mutual authentication in Kyma and make sure that you exported the [bundle certificates](../00-security/sec-02-setup-mtls-gateway#steps).

You can also take a look on tutorial [How to create own self-signed Client Root CA and Certificate](../00-security/sec-02-mtls-selfsign-client-certicate.md)

## Authorize client with a certificate

Follow the instructions in the tabs to further secure the mTLS service or Function. Create AuthorizationPolicy that checks if the client's Common Name in the certificate matches.

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
     - ${GATEWAY}
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

2. Create AuthorizationPolicy that verifies if the request contains a new client certificate:
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
     - "function-example.${DOMAIN_TO_EXPOSE_WORKLOADS}"
     gateways:
     - ${GATEWAY}
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
2. Create AuthorizationPolicy that verifies if the request contains a new client certificate:
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
           hosts: ["function-example.${DOMAIN_TO_EXPOSE_WORKLOADS}"]
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
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \ -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
   ```

These calls return the code `200` response. If you call the service without the proper certificates or with old ones, you get the code `403` response.

  </details>

  <details>
  <summary>
  Call the secured Function
  </summary>

Send a `GET` request with a token that has the `read` scope to the Function:

   ```shell
   curl --key ${CLIENT_CERT_KEY_FILE} \
        --cert ${CLIENT_CERT_CRT_FILE} \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \ -ik -X GET https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```

This call returns the code `200` response. If you call the Function without the proper certificates or with old ones, you get the code `403` response.
  </details>
</div>