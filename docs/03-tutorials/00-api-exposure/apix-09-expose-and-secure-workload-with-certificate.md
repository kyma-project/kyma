---
title: Expose and secure a workload with a certificate 
---

This tutorial shows how to expose a workload with mutual authentication using Kyma's mutual TLS Gateway. 

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

Before you start, Set up [`kyma-mtls-gateway`](../00-security/sec-02-setup-mtls-gateway.md) to allow mutual authentication in Kyma, and make sure that you exported the bundle certificates. 

## Expose and access your workload

Follow the instruction to expose and access your instance of the HttpBin service or your sample Function.

<div tabs>
  <details>
  <summary>
  HttpBin
  </summary>

1. Expose unsecured workload on mTLS Gateway
   ```bash
   cat <<EOF | kubectl apply -f -
   ---
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: httpbin-mtls-gw-unsecured
     namespace: ${MTLS_TEST_NAMESPACE}
   spec:
     host: httpbin-vs.${CUSTOM_DOMAIN}
     service:
       name: httpbin
       port: 8000
     gateway: ${MTLS_TEST_NAMESPACE}/${MTLS_GATEWAY_NAME}
     rules:
       - path: /.*
         methods: ["GET"]
         accessStrategies:
           - handler: noop
         mutators:
           - handler: noop
       - path: /post
         methods: ["POST"]
         accessStrategies:
           - handler: noop
         mutators:
           - handler: noop
   EOF
   ```
2. Verify if the workload is accessible
   ```bash
   curl --key ${CLIENT_CERT_KEY_FILE} \
        --cert ${CLIENT_CERT_CRT_FILE} \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
        -ik -X GET https://httpbin.${CUSTOM_DOMAIN}/ip
   ```
3. Create Virtual service
   ```bash
   cat <<EOF | kubectl apply -f - 
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: httpbin-vs
     namespace: ${MTLS_TEST_NAMESPACE}
   spec:
     hosts:
     - "httpbin-vs.${CUSTOM_DOMAIN}"
     gateways:
     - ${MTLS_TEST_NAMESPACE}/${MTLS_GATEWAY_NAME}
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
4. Create AuthorizationPolicy
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: security.istio.io/v1beta1
   kind: AuthorizationPolicy
   metadata:
     name: test-authz-policy
     namespace: ${MTLS_TEST_NAMESPACE}
   spec:
     action: ALLOW
     rules:
     - to:
       - operation:
           hosts: ["httpbin-vs.mtls-gw.goat.build.kyma-project.io"]
       when:
       - key: request.headers[X-Client-Ssl-Cn]
         values: ["O=example,CN=client2.example.com"]
   EOF
   ```
  </details>
  <details>
  <summary>
  Function
  </summary>

1. Expose unsecured workload on mTLS Gateway
   ```bash
   cat <<EOF | kubectl apply -f -
   ---
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: function-mtls-gw-unsecured
     namespace: ${MTLS_TEST_NAMESPACE}
   spec:
     host: function-example.${CUSTOM_DOMAIN}
     service:
       name: function
       port: 80
     gateway: ${MTLS_TEST_NAMESPACE}/${MTLS_GATEWAY_NAME}
     rules:
       - path: /.*
         methods: ["GET"]
         accessStrategies:
           - handler: noop
         mutators:
           - handler: noop
       - path: /post
         methods: ["POST"]
         accessStrategies:
           - handler: noop
         mutators:
           - handler: noop
   EOF
   ```
2. Verify if the workload is accessible
   ```bash
   curl --key ${CLIENT_CERT_KEY_FILE} \
        --cert ${CLIENT_CERT_CRT_FILE} \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
        -ik -X GET https://function-example.${CUSTOM_DOMAIN}/function
   ```
3. Create Virtual service
   ```bash
   cat <<EOF | kubectl apply -f - 
   apiVersion: networking.istio.io/v1alpha3
   kind: VirtualService
   metadata:
     name: function-vs
     namespace: ${MTLS_TEST_NAMESPACE}
   spec:
     hosts:
     - "function-example.${CUSTOM_DOMAIN}"
     gateways:
     - ${MTLS_TEST_NAMESPACE}/${MTLS_GATEWAY_NAME}
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
4. Create AuthorizationPolicy
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: security.istio.io/v1beta1
   kind: AuthorizationPolicy
   metadata:
     name: test-authz-policy
     namespace: ${MTLS_TEST_NAMESPACE}
   spec:
     action: ALLOW
     rules:
     - to:
       - operation:
           hosts: ["function-example.${CUSTOM_DOMAIN}"]
       when:
       - key: request.headers[X-Client-Ssl-Cn]
         values: ["O=example,CN=client2.example.com"]
   EOF
   ```
  </details>
</div>


## Access the secured resources

Follow the instructions in the tabs to call the secured service or Functions using the certificates for the mTLS Gateway.

<div tabs>

  <details>
  <summary>
  Call secured endpoints of a service
  </summary>

1. Send a `GET` request to the HttpBin service with a client certificates that were used to create mTLS Gateway:

   ```shell
   curl --key ${CLIENT_CERT_KEY_FILE} \
        --cert ${CLIENT_CERT_CRT_FILE} \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \ -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
   ```

These calls return the code `200` response. If you call the service without proper certificates, you get the code `403` response.

  </details>

  <details>
  <summary>
  Call the secured Function
  </summary>

Send a `GET` request with a token that has the "read" scope to the Function:

   ```shell
   curl --key ${CLIENT_CERT_KEY_FILE} \
        --cert ${CLIENT_CERT_CRT_FILE} \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \ -ik -X GET https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```

This call returns the code `200` response. If you call the Function without without proper certificates, you get the code `403` response.
  </details>
</div>


