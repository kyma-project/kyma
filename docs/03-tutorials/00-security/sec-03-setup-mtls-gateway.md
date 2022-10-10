---
title: Set up a mutual TLS Gateway
---

This tutorial shows how to set up a mutual TLS Gateway and configure authentication based on certificate details.

## Prerequisites

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

3. Export Gateway name, TLS secret name, and client root Certificate Authority (CA) crt file path:

    ```bash
   export MTLS_GATEWAY_NAME=mtls-gateway
   export TLS_SECRET={TLS_SECRET_NAME} # The name of the TLS Secret that was created during the setup of the custom domain, for example, httpbin-tls-credentials
   export CLIENT_ROOT_CA_CRT_FILE={CLIENT_ROOT_CA_CRT_FILE}
   ```

4. Create mTLS Gateway (mode: MUTUAL)
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

5. Add Client Root CA to cacert bundle secret for mTLS Gateway. Export the following value as an environment variable and run the command provided:

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
