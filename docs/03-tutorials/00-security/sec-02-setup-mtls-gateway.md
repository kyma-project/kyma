---
title: Set up a mutual TLS Gateway 
---

This tutorial shows how to set up mutual TLS Gateway and configure authentication based on certificate details.

## Prerequisites

To follow this tutorial, use Kyma 2.6 or higher.

If you use a cluster not managed by Gardener, install the [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components manually in a dedicated Namespace.

## Steps

1. Export the following values as environment variables:

   ```bash
   # CLIENT Certificates
   export CLIENT_ROOT_CA_KEY_FILE=client-root-ca.key
   export CLIENT_ROOT_CA_CRT_FILE=client-root-ca.crt
   export CLIENT_CERT_CRT_FILE=client.example.com.crt
   export CLIENT_CERT_CSR_FILE=client.example.com.csr
   export CLIENT_CERT_KEY_FILE=client.example.com.key 
   # Domain
   export DOMAIN=goat.build.kyma-project.io
   export CUSTOM_DOMAIN=mtls-gw.$DOMAIN
   # CRDs variables 
   export MTLS_CERT_SECRET_NAME=mtls-gw-cert
   export MTLS_GATEWAY_NAME=mtls-gateway
   export MTLS_TEST_NAMESPACE=mtls-test
   ```

2. Create Namespace

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Namespace
    metadata:
      name: ${MTLS_TEST_NAMESPACE}
      labels: 
        istio-injection: enabled
    EOF
    ```
   
3. Manage DNS
    ```bash
    export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    ```
    
4. Check if loadbalancer uses IP
    
    ```bash
    echo "IP: <<" $IP ">> if empty use HOSTNAME instead of IP"
    ```
       >**NOTE:** If the previous output is empty, use HOSTNAME instead of IP

    ```bash
    export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    ```
    
5. Create DNSEntry for CUSTOM_DOMAIN
    ```bash
    cat <<EOF | kubectl apply -f -  
    --- 
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSEntry
    metadata:
      name: dns-entry
      namespace: ${MTLS_TEST_NAMESPACE}
      annotations:
        dns.gardener.cloud/class: garden
    spec:
      dnsName: "*.${CUSTOM_DOMAIN}"
      ttl: 600
      targets:
        - $IP
   EOF
   ```
   
6. Manage server TLS
   ```bash
   cat <<EOF | kubectl apply -f -  
   --- 
   apiVersion: cert.gardener.cloud/v1alpha1
   kind: Certificate
   metadata:
     name: ${MTLS_CERT_SECRET_NAME}
     namespace: istio-system
   spec:
     commonName: "*.${CUSTOM_DOMAIN}"
     issuerRef:
       name: garden
     secretRef:
       name: ${MTLS_CERT_SECRET_NAME} 
       namespace: istio-system
   EOF
   ```
7. Manage Client TLS
   Create Client Root CA and CLient certificate signed by Client Root CA
   ```bash
   # Create root ca key and cert (valid for 5 years) - will be used for validation
   openssl req -x509 -sha256 -nodes -days 1825 -newkey rsa:2048 -subj '/O=example Inc./CN=ClientRootCA' -keyout ${CLIENT_ROOT_CA_KEY_FILE} -out ${CLIENT_ROOT_CA_CRT_FILE}
   # Create new key and csr for client cert
   openssl req -out ${CLIENT_CERT_CSR_FILE} -newkey rsa:2048 -nodes -keyout ${CLIENT_CERT_KEY_FILE} -subj "/CN=client.example.com/O=example"
   # Sign client cert with CA cert
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${CLIENT_CERT_CSR_FILE} -out ${CLIENT_CERT_CRT_FILE}
   # Create new key and csr for client cert
   openssl req -out ${CLIENT_CERT_CSR_FILE}.2 -newkey rsa:2048 -nodes -keyout ${CLIENT_CERT_KEY_FILE}.2 -subj "/CN=client2.example.com/O=example"
   # Sign client cert with CA cert
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${CLIENT_CERT_CSR_FILE}.2 -out ${CLIENT_CERT_CRT_FILE}.2
   
   ```

8. Add Client Root CA to cacert bundle secret for mTLS Gateway
   ```bash
   # Add CA Cert to kyma-mtls-gateway
   export CLIENT_ROOT_CA_CRT_ENCODED=$(cat ${CLIENT_ROOT_CA_CRT_FILE}| base64)
   cat <<EOF | kubectl apply -f -
   ---
   apiVersion: v1
   kind: Secret
   metadata:
     name: ${MTLS_CERT_SECRET_NAME}-cacert
     namespace: istio-system
   type: Opaque
   data:
     cacert: ${CLIENT_ROOT_CA_CRT_ENCODED}
   EOF
   ```

9. Create mTLS Gateway (mode: MUTUAL)
   ```bash
   cat <<EOF | kubectl apply -f -
   ---
   apiVersion: networking.istio.io/v1beta1
   kind: Gateway
   metadata:
     name: ${MTLS_GATEWAY_NAME}
     namespace: ${MTLS_TEST_NAMESPACE}
   spec:
     selector:
       istio: ingressgateway
       app: istio-ingressgateway
     servers:
       - port:
           number: 443
           name: https
           protocol: HTTPS
         tls:
           mode: MUTUAL
           credentialName: ${MTLS_CERT_SECRET_NAME}
           minProtocolVersion: TLSV1_2
           cipherSuites:
           - ECDHE-RSA-CHACHA20-POLY1305
           - ECDHE-RSA-AES256-GCM-SHA384
           - ECDHE-RSA-AES256-SHA
           - ECDHE-RSA-AES128-GCM-SHA256
           - ECDHE-RSA-AES128-SHA
         hosts:
           - '*.${CUSTOM_DOMAIN}'
       - port:
           number: 80
           name: http
           protocol: HTTP
         tls:
           httpsRedirect: true
         hosts:
           - '*.${CUSTOM_DOMAIN}'
   EOF
   ```