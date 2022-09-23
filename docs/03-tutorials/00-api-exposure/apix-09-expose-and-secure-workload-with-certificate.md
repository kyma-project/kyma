---
title: Expose and secure a workload with a certificate 
---

This tutorial shows how to expose and secure a workload with mutual authentication using Kyma's mutual TLS Gateway. 

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

Before you start, Set up [`kyma-mtls-gateway`](../00-security/sec-02-setup-mtls-gateway.md) to allow mutual authentication in Kyma and make sure that you exported the bundle certificates. 

## Expose and access your workload

Follow the instruction to expose and access your instance of the HttpBin service or your sample Function.

<div tabs>
  <details>
  <summary>
  HttpBin
  </summary>

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/$MTLS_GATEWAY_NAME
   export CLIENT_CERT_CRT_FILE=client.example.com.crt
	 export CLIENT_CERT_CSR_FILE=client.example.com.csr
	 export CLIENT_CERT_KEY_FILE=client.example.com.key 
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com.

2. Expose the instance of the HttpBin service on mTLS Gateway by creating an APIRule CR in your Namespace. Run:
   ```bash
   cat <<EOF | kubectl apply -f -
   ---
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: httpbin-mtls-gw-unsecured
     namespace: ${NAMESPACE}
   spec:
     host: httpbin-vs.${DOMAIN_TO_EXPOSE_WORKLOADS}
     service:
       name: httpbin
       port: 8000
     gateway: ${GATEWAY}
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

3. Generate Client certificate signed by Client Root CA: 

   ```bash
   # Create a new key and CSR for the client certificate
   openssl req -out ${CLIENT_CERT_CSR_FILE} -newkey rsa:2048 -nodes -keyout ${CLIENT_CERT_KEY_FILE} -subj "/CN=client.example.com/O=example"
   # Sign the client cert with CA cert
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${CLIENT_CERT_CSR_FILE} -out ${CLIENT_CERT_CRT_FILE}
   ```
  
  </details>
  <details>
  <summary>
  Function
  </summary>
  
1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE/$MTLS_GATEWAY_NAME
   export CLIENT_CERT_CRT_FILE=client.example.com.crt
	 export CLIENT_CERT_CSR_FILE=client.example.com.csr
	 export CLIENT_CERT_KEY_FILE=client.example.com.key 
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com

2. Expose the sample Function on mTLS Gateway by creating an APIRule CR in your Namespace. Run:
   ```bash
   cat <<EOF | kubectl apply -f -
   ---
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: function-mtls-gw-unsecured
     namespace: ${NAMESPACE}
   spec:
     host: function-example.${DOMAIN_TO_EXPOSE_WORKLOADS}
     service:
       name: function
       port: 80
     gateway: ${GATEWAY}
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
3. Generate Client certificate signed by Client Root CA: 

   ```bash
   # Create a new key and CSR for the client certificate
   openssl req -out ${CLIENT_CERT_CSR_FILE} -newkey rsa:2048 -nodes -keyout ${CLIENT_CERT_KEY_FILE} -subj "/CN=client.example.com/O=example"
   # Sign the client cert with CA cert
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${CLIENT_CERT_CSR_FILE} -out ${CLIENT_CERT_CRT_FILE}
   ```

  </details>
</div>

## Access the secured resources

Follow the instructions in the tabs to call the secured service or Function using the certificates for the mTLS Gateway.

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

These calls return the code `200` response. If you call the service without the proper certificates, you get the code `403` response.

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

This call returns the code `200` response. If you call the Function without the proper certificates, you get the code `403` response.
  </details>
</div>

## Further secure the mTLS endpoint

Follow the instructions in the tabs to further secure the mTLS service or Function, by creating AuthorizationPolicy that will check if client's Common Name in certificate matches.

<div tabs>
  <details>
  <summary>
  HttpBin
  </summary>

1. Export the following value as an environment variable:

   ```bash
	 export NEW_CLIENT_CERT_CRT_FILE=client2.example.com.crt
	 export NEW_CLIENT_CERT_CSR_FILE=client2.example.com.csr
	 export NEW_CLIENT_CERT_KEY_FILE=client2.example.com.key 
   ```

2. Generate a new Client certificate signed by Client Root CA

   ```bash
   # Create a new key and CSR for the client certificate
   openssl req -out ${NEW_CLIENT_CERT_CSR_FILE} -newkey rsa:2048 -nodes -keyout ${NEW_CLIENT_CERT_KEY_FILE} -subj "/CN=client2.example.com/O=example"
   # Sign the client certificate with CA cert
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${NEW_CLIENT_CERT_CSR_FILE} -out ${NEW_CLIENT_CERT_CRT_FILE}
   ```


3. Create VirtualService that will add X-CLIENT-SSL headers to incoming requests:
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

4. Create AuthorizationPolicy that will verify if the request contains a new client certificate:
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
         values: ["O=example,CN=client2.example.com"]
   EOF
   ```
  </details>
  <details>
  <summary>
  Function
  </summary>
  
1. Export the following value as an environment variable:

   ```bash
	 export NEW_CLIENT_CERT_CRT_FILE=client2.example.com.crt
	 export NEW_CLIENT_CERT_CSR_FILE=client2.example.com.csr
	 export NEW_CLIENT_CERT_KEY_FILE=client2.example.com.key 
   ```

2. Generate a new Client certificate signed by Client Root CA

   ```bash
   # Create a new key and CSR for the client certificate
   openssl req -out ${NEW_CLIENT_CERT_CSR_FILE} -newkey rsa:2048 -nodes -keyout ${NEW_CLIENT_CERT_KEY_FILE} -subj "/CN=client2.example.com/O=example"
   # Sign the client certificate with CA cert
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${NEW_CLIENT_CERT_CSR_FILE} -out ${NEW_CLIENT_CERT_CRT_FILE}
   ```

3. Create VirtualService that will add X-CLIENT-SSL headers to incoming requests:
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
4. Create AuthorizationPolicy that will verify if the request contains a new client certificate:
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
         values: ["O=example,CN=client2.example.com"]
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
   curl --key ${NEW_CLIENT_CERT_KEY_FILE} \
        --cert ${NEW_CLIENT_CERT_CRT_FILE} \
        --cacert ${NEW_CLIENT_ROOT_CA_CRT_FILE} \ -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
   ```

These calls return the code `200` response. If you call the service without the proper certificates or with old ones, you get the code `403` response.

  </details>

  <details>
  <summary>
  Call the secured Function
  </summary>

Send a `GET` request with a token that has the "read" scope to the Function:

   ```shell
   curl --key ${NEW_CLIENT_CERT_KEY_FILE} \
        --cert ${NEW_CLIENT_CERT_CRT_FILE} \
        --cacert ${NEW_CLIENT_ROOT_CA_CRT_FILE} \ -ik -X GET https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```

This call returns the code `200` response. If you call the Function without the proper certificates or with old ones, you get the code `403` response.
  </details>
</div>
