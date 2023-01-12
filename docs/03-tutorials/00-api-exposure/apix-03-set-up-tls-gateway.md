---
title: Set up a TLS Gateway
---

This tutorial shows how to set up a TLS Gateway in both manual and simple modes. It also explains how to configure authentication for an mTLS Gateway based on certificate details.

## Prerequisites

* [Custom domain](../00-api-exposure/apix-02-setup-custom-domain-for-workload.md) set up

## Set up a TLS Gateway

1. Export the following values as environment variables:

   ```bash
   export NAMESPACE={NAMESPACE_NAME}
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} 
   export TLS_SECRET={TLS_SECRET_NAME}
   ```
   >**NOTE:** The `DOMAIN_NAME` refers to the domain that you own, for example, `api.mydomain.com`. The `TLS_SECRET` is the name of the TLS Secret created during the setup of the custom domain, for example, `httpbin-tls-credentials`. 

2. Create a TLS Gateway.

  Follow the instructions to create a TLS Gateway either in simple or in mutual mode.
  
  <div tabs>

    <details>
    <summary>
    simple mode
    </summary>

    * To create a TLS Gateway in simple mode, run:

    ```bash
      cat <<EOF | kubectl apply -f -
      apiVersion: networking.istio.io/v1alpha3
      kind: Gateway
      metadata:
        name: httpbin-gateway
        namespace: $NAMESPACE
      spec:
        selector:
          istio: ingressgateway # Use Istio Ingress Gateway as default
        servers:
          - port:
              number: 443
              name: https
              protocol: HTTPS
            tls:
              mode: SIMPLE
              credentialName: $TLS_SECRET
            hosts:
              - "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
      EOF
    ```
    
    </details>

    <details>
    <summary>
    mutual mode
    </summary>

    * Export the name of the mTLS Gateway and the client root Certificate Authority (CA) crt file path as environment variables:

      ```bash
      export MTLS_GATEWAY_NAME=mtls-gateway
      export CLIENT_ROOT_CA_CRT_FILE={CLIENT_ROOT_CA_CRT_FILE}

    * To create a mutual TLS Gateway, run:
    
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
    * Export the following value as an environment variable:

      ```bash
      export CLIENT_ROOT_CA_CRT_ENCODED=$(cat ${CLIENT_ROOT_CA_CRT_FILE}| base64)
      ```

    * Add client root CA to the CA cert bundle Secret for mTLS Gateway. Run:

    ```bash
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
    </details>
  </div>