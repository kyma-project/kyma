---
title: Expose and secure a workload with a certificate 
---

This tutorial shows how to expose a workload with mutual authentication using Kyma's mutual TLS Gateway. 

The tutorial may be a follow-up to the [Set up a custom domain for a workload](./apix-02-setup-custom-domain-for-workload.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

Before you start, you must set up [`mtls-gateway`](../00-security/sec-02-setup-mtls-gateway.md) to allow mutual authentication in Kyma. 

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
     host: httpbin.${CUSTOM_DOMAIN}
     service:
       name: httpbin
       port: 8000
     gateway: ${MTLS_GATEWAY_NAME}.${MTLS_TEST_NAMESPACE}.svc.cluster.local
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
5. Test
   Allowed cert
   ```bash
   curl --key ${CLIENT_CERT_KEY_FILE}.2 \
        --cert ${CLIENT_CERT_CRT_FILE}.2 \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
        -ik -X GET https://httpbin-vs.${CUSTOM_DOMAIN}/headers
   ```
   
   ```bash
   HTTP/2 200 
   server: istio-envoy
   date: Thu, 08 Sep 2022 06:26:21 GMT
   content-type: application/json
   content-length: 2005
   access-control-allow-origin: *
   access-control-allow-credentials: true
   x-envoy-upstream-service-time: 6
   
   {
     "headers": {
       "Accept": "*/*", 
       "Host": "httpbin-vs.mtls-gw.goat.build.kyma-project.io", 
       "User-Agent": "curl/7.79.1", 
       "X-B3-Parentspanid": "630ed85f411a73e3", 
       "X-B3-Sampled": "0", 
       "X-B3-Spanid": "03c71df9d75f54ee", 
       "X-B3-Traceid": "325d5730877962ff630ed85f411a73e3", 
       "X-Client-Ssl-Cn": "O=example,CN=client2.example.com", 
       "X-Client-Ssl-Issuer": "CN=ClientRootCA,O=example Inc.", 
       "X-Envoy-Attempt-Count": "1", 
       "X-Envoy-Internal": "true", 
       "X-Forwarded-Client-Cert": "Hash=269d8c80411226e3d867699542b860d030da826aa5172a402572b486d04e31e8;Cert=\"-----BEGIN%20CERTIFICATE-----%0AMIIC0jCCAboCAQAwDQYJKoZIhvcNAQEFBQAwLjEVMBMGA1UECgwMZXhhbXBsZSBJ%0AbmMuMRUwEwYDVQQDDAxDbGllbnRSb290Q0EwHhcNMjIwOTA3MTMzNDE3WhcNMjMw%0AOTA3MTMzNDE3WjAwMRwwGgYDVQQDDBNjbGllbnQyLmV4YW1wbGUuY29tMRAwDgYD%0AVQQKDAdleGFtcGxlMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArTOS%0Aom8hybqygtmI55ogWX5FqOd%2Byy%2BaknRvs%2B58YiJFzd5kc2hRcys0ZaDIF2feOyxv%0AqCImIfxTO%2FJniGVw%2BqzymffeniBkUna7wnHJbncazG%2FH2MvKV31hVdi2BzNeoLhy%0AsFlAaVy0Ernl2CJgIdiDrd0iHuC%2FCsHwUanWrIRkVz3W3CDYx0b%2Brfe8xwj3unSD%0AGTRE7h41EcXtvkDt0hcsjdM7TXKYwrozif6h9mbvM8ZYFhQ%2BpEI9hsrNbnzA4bFM%0AEKIi%2Bze18EK0lqnjUX96Aipb8mav3cpaz4ZYzl2M5wbwY1jYhz5YudmovIniO0CM%0AqyKChkX1BqcSN0w3WQIDAQABMA0GCSqGSIb3DQEBBQUAA4IBAQApGZhv5KruCsID%0A9Q5SCiA2oSyCTkpfDS%2BJn3hJMYEnfk3WtiMVJriU2El0TA1jrnuDg%2FMnMPYm8rmj%0AyJgBktXWTStqtocz8CrOnEAHtj%2FHCrkIzvYhj23sUrUspFZ2loHjH14n%2Bsf9eUbt%0Ax9ofHmph%2BpmY%2BAh96gG8dervfCxY4WveHKboh1FIsbSv2Am%2BhUccUErWTM8Tyfdh%0A%2F%2BB4hi%2BCZUZNYfR%2B5pFcAMc7b7AQpcLq4G3KZiAV03ppx1A%2FptBN6eH9X2CRfLRk%0A8j9IlYXpmTuPFG8bF1yjFWlVQZxmYO%2FhW2Oo%2BMmyUFWkrjqEZG61DCL1BzpOGfsb%0AJO1hCIg5%0A-----END%20CERTIFICATE-----%0A\";Subject=\"O=example,CN=client2.example.com\";URI=,By=spiffe://cluster.local/ns/mtls-test/sa/httpbin;Hash=eb7bc6749d2b65c19ae0600fee7a9e34fbe1d24527efc78180448361da02aaef;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"
     }
   }
   ```
   Denied cert
   ```bash
   curl --key ${CLIENT_CERT_KEY_FILE} \  
     --cert ${CLIENT_CERT_CRT_FILE} \  
     --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
     -ik -X GET https://httpbin-vs.${CUSTOM_DOMAIN}/headers
   ```
   ```bash
   HTTP/2 403 
   content-length: 19
   content-type: text/plain
   date: Thu, 08 Sep 2022 06:27:06 GMT
   server: istio-envoy
   x-envoy-upstream-service-time: 1
   
   RBAC: access denied%  
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
5. Test
   Allowed cert
   ```bash
   curl --key ${CLIENT_CERT_KEY_FILE}.2 \
        --cert ${CLIENT_CERT_CRT_FILE}.2 \
        --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
        -ik -X GET https://function-example.${CUSTOM_DOMAIN}/function
   ```
   
   ```bash
   HTTP/2 200 
   server: istio-envoy
   date: Thu, 08 Sep 2022 06:26:21 GMT
   content-type: application/json
   content-length: 2005
   access-control-allow-origin: *
   access-control-allow-credentials: true
   x-envoy-upstream-service-time: 6
   
   {
     "headers": {
       "Accept": "*/*", 
       "Host": "httpbin-vs.mtls-gw.goat.build.kyma-project.io", 
       "User-Agent": "curl/7.79.1", 
       "X-B3-Parentspanid": "630ed85f411a73e3", 
       "X-B3-Sampled": "0", 
       "X-B3-Spanid": "03c71df9d75f54ee", 
       "X-B3-Traceid": "325d5730877962ff630ed85f411a73e3", 
       "X-Client-Ssl-Cn": "O=example,CN=client2.example.com", 
       "X-Client-Ssl-Issuer": "CN=ClientRootCA,O=example Inc.", 
       "X-Envoy-Attempt-Count": "1", 
       "X-Envoy-Internal": "true", 
       "X-Forwarded-Client-Cert": "Hash=269d8c80411226e3d867699542b860d030da826aa5172a402572b486d04e31e8;Cert=\"-----BEGIN%20CERTIFICATE-----%0AMIIC0jCCAboCAQAwDQYJKoZIhvcNAQEFBQAwLjEVMBMGA1UECgwMZXhhbXBsZSBJ%0AbmMuMRUwEwYDVQQDDAxDbGllbnRSb290Q0EwHhcNMjIwOTA3MTMzNDE3WhcNMjMw%0AOTA3MTMzNDE3WjAwMRwwGgYDVQQDDBNjbGllbnQyLmV4YW1wbGUuY29tMRAwDgYD%0AVQQKDAdleGFtcGxlMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArTOS%0Aom8hybqygtmI55ogWX5FqOd%2Byy%2BaknRvs%2B58YiJFzd5kc2hRcys0ZaDIF2feOyxv%0AqCImIfxTO%2FJniGVw%2BqzymffeniBkUna7wnHJbncazG%2FH2MvKV31hVdi2BzNeoLhy%0AsFlAaVy0Ernl2CJgIdiDrd0iHuC%2FCsHwUanWrIRkVz3W3CDYx0b%2Brfe8xwj3unSD%0AGTRE7h41EcXtvkDt0hcsjdM7TXKYwrozif6h9mbvM8ZYFhQ%2BpEI9hsrNbnzA4bFM%0AEKIi%2Bze18EK0lqnjUX96Aipb8mav3cpaz4ZYzl2M5wbwY1jYhz5YudmovIniO0CM%0AqyKChkX1BqcSN0w3WQIDAQABMA0GCSqGSIb3DQEBBQUAA4IBAQApGZhv5KruCsID%0A9Q5SCiA2oSyCTkpfDS%2BJn3hJMYEnfk3WtiMVJriU2El0TA1jrnuDg%2FMnMPYm8rmj%0AyJgBktXWTStqtocz8CrOnEAHtj%2FHCrkIzvYhj23sUrUspFZ2loHjH14n%2Bsf9eUbt%0Ax9ofHmph%2BpmY%2BAh96gG8dervfCxY4WveHKboh1FIsbSv2Am%2BhUccUErWTM8Tyfdh%0A%2F%2BB4hi%2BCZUZNYfR%2B5pFcAMc7b7AQpcLq4G3KZiAV03ppx1A%2FptBN6eH9X2CRfLRk%0A8j9IlYXpmTuPFG8bF1yjFWlVQZxmYO%2FhW2Oo%2BMmyUFWkrjqEZG61DCL1BzpOGfsb%0AJO1hCIg5%0A-----END%20CERTIFICATE-----%0A\";Subject=\"O=example,CN=client2.example.com\";URI=,By=spiffe://cluster.local/ns/mtls-test/sa/httpbin;Hash=eb7bc6749d2b65c19ae0600fee7a9e34fbe1d24527efc78180448361da02aaef;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"
     }
   }
   ```
   Denied cert
   ```bash
   curl --key ${CLIENT_CERT_KEY_FILE} \  
     --cert ${CLIENT_CERT_CRT_FILE} \  
     --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
     -ik -X GET https://function-example.${CUSTOM_DOMAIN}/function
   ```
   ```bash
   HTTP/2 403 
   content-length: 19
   content-type: text/plain
   date: Thu, 08 Sep 2022 06:27:06 GMT
   server: istio-envoy
   x-envoy-upstream-service-time: 1
   
   RBAC: access denied%  
   ```
  </details>
</div>









# Related links
- https://istio.io/latest/docs/reference/config/security/conditions/
- https://istio.io/latest/docs/reference/config/security/authorization-policy/#Condition
- https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers.html?highlight=forward_client_cert
- https://istio.io/latest/docs/ops/configuration/traffic-management/network-topologies/
