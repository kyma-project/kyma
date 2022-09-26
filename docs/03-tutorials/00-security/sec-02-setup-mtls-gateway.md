---
title: Set up a mutual TLS Gateway 
---

This tutorial shows how to set up a mutual TLS Gateway and configure authentication based on certificate details.

## Prerequisites

To follow this tutorial, use Kyma 2.6 or higher.

Before you start, Set up a [`custom-domain`](../00-api-exposure/apix-02-setup-custom-domain-for-workload.md) and prepare a certificate to expose a workload.

## Steps

1. Export the following values as environment variables:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} 
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com

2. Create a Namespace and export its value as an environment variable. Run:
   >**NOTE:** Skip this step if you already have a Namespace

   ```bash
    export NAMESPACE={NAMESPACE_NAME}
	  kubectl create ns $NAMESPACE
	  kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
   ```

3. Create Client Root CA and CLient certificate signed by Client Root CA. Export the following values as environment variables and run the command provided:  
   ```bash
    export CLIENT_ROOT_CA_KEY_FILE=client-root-ca.key
	  export CLIENT_ROOT_CA_CRT_FILE=client-root-ca.crt
	  export CLIENT_CERT_CRT_FILE=client.example.com.crt
	  export CLIENT_CERT_CSR_FILE=client.example.com.csr
	  export CLIENT_CERT_KEY_FILE=client.example.com.key 
   ```

   ```bash
   # Create root CA key and cert (valid for 5 years) - will be used for validation
   openssl req -x509 -sha256 -nodes -days 1825 -newkey rsa:2048 -subj '/O=example Inc./CN=ClientRootCA' -keyout ${CLIENT_ROOT_CA_KEY_FILE} -out ${CLIENT_ROOT_CA_CRT_FILE}
   # Create a new key and CSR for the client certificate
   openssl req -out ${CLIENT_CERT_CSR_FILE} -newkey rsa:2048 -nodes -keyout ${CLIENT_CERT_KEY_FILE} -subj "/CN=client.example.com/O=example"
   # Sign the client cert with CA cert
   openssl x509 -req -days 365 -CA ${CLIENT_ROOT_CA_CRT_FILE} -CAkey ${CLIENT_ROOT_CA_KEY_FILE} -set_serial 0 -in ${CLIENT_CERT_CSR_FILE} -out ${CLIENT_CERT_CRT_FILE}
   ```

4. Add Client Root CA to cacert bundle secret for mTLS Gateway. Export the following value as an environment variable and run the command provided:

   ```bash
   export MTLS_GATEWAY_NAME=mtls-gateway
   export TLS_SECRET={TLS_SECRET_NAME} # The name of the TLS Secret that was created during the setup of the custom domain, for example, httpbin-tls-credentials
   ```

   ```bash
   # Add CA Cert to kyma-mtls-gateway
   export CLIENT_ROOT_CA_CRT_ENCODED=$(cat ${CLIENT_ROOT_CA_CRT_FILE}| base64)
   cat <<EOF | kubectl apply -f -
   ---
   apiVersion: v1
   kind: Secret
   metadata:
     name: ${TLS_SECRET}-cacert
     namespace: istio-system
   type: Opaque
   data:
     cacert: ${CLIENT_ROOT_CA_CRT_ENCODED}
   EOF
   ```

5. Create mTLS Gateway (mode: MUTUAL)
   ```bash
   cat <<EOF | kubectl apply -f -
   ---
   apiVersion: networking.istio.io/v1beta1
   kind: Gateway
   metadata:
     name: ${MTLS_GATEWAY_NAME}
     namespace: ${NAMESPACE}
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
           credentialName: ${TLS_SECRET}
           minProtocolVersion: TLSV1_2
           cipherSuites:
           - ECDHE-RSA-CHACHA20-POLY1305
           - ECDHE-RSA-AES256-GCM-SHA384
           - ECDHE-RSA-AES256-SHA
           - ECDHE-RSA-AES128-GCM-SHA256
           - ECDHE-RSA-AES128-SHA
         hosts:
           - '*.${DOMAIN_TO_EXPOSE_WORKLOADS}'
       - port:
           number: 80
           name: http
           protocol: HTTP
         tls:
           httpsRedirect: true
         hosts:
           - '*.${DOMAIN_TO_EXPOSE_WORKLOADS}'
   EOF
   ```