---
title: Update TLS certificate
type: Details
---

Once in a while it is necessary to change server TLS certificates, usually because they got expired. Follow the 
instruction to Update TLS certificates in Kyma. 

## Prerequisites
 * New TLS certificates
 * Admin access to kyma 

**This procedure may cause short interruption in access to Kyma from outside the cluster.**

## Update certificates

1. Export certificates. Replace `/path/to/...` with paths to your certificates. 

    ```bash
    export KYMA_TLS_CERT=$(cat /path/to/cert.pem)
    export KYMA_TLS_KEY=$(cat /path/to/key.pem)
    ```

2. Update istio ingressgateway certificate.

    ```bash
    cat <<EOF | kubectl create -f -
    apiVersion: v1
    kind: Secret
    metadata:
        name: istio-ingressgateway-cert
        namespace: istio-system
    data:
        tls.crt: $(echo "${KYMA_TLS_CERT}" | base64)
        tls.key: $(echo "${KYMA_TLS_KEY}" | base64)
    EOF
    ```
 
3. Update kyma-system certificate.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Secret
    metadata:
        name: ingress-tls-cert
        namespace: kyma-system
    data:
        tls.crt: $(echo "${KYMA_TLS_CERT}" | base64)
        tls.key: $(echo "${KYMA_TLS_KEY}" | base64)
    EOF
    ```
    
4. Update kyma-integration certificate.

    ```bash
    cat <<EOF | kubectl create -f -
    apiVersion: v1
    kind: Secret
    metadata:
        name: ingress-tls-cert
        namespace: kyma-integration
    data:
        tls.crt: $(echo "${KYMA_TLS_CERT}" | base64)
        tls.key: $(echo "${KYMA_TLS_KEY}" | base64)
    EOF
    ```

5. Restart istio-ingressgateway pod to pick up new certificate. 

    ```bash
    kubectl delete pod -l app=istio-ingressgateway -n istio-system
    ```
    
6. Restart pods in kyma-system pod to pick up new certificate. 

    ```bash
    kubectl delete pod -l tlsSecret=ingress-tls-cert -n kyma-system
    ```
    
7. Restart pods in kyma-integration pod to pick up new certificate. 

    ```bash
    kubectl delete pod -l tlsSecret=ingress-tls-cert -n kyma-integration
    ```