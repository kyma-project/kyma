---
title: Update TLS certificate
type: Details
---

The TLS certificate is a vital security element. This document describes how to update the TLS certificate in Kyma.

>**NOTE:** This procedure can interrupt the communication between your cluster and the outside world for a limited 
period of time.

## Prerequisites
 * New TLS certificates
 * Kyma administrator access 

## Steps

1. Export the new TLS certificate and key as environment variables. Run:

    ```bash
    export KYMA_TLS_CERT=$(cat {NEW_CERT_PATH})
    export KYMA_TLS_KEY=$(cat {NEW_KEY_PATH})
    ```

2. Update the Ingress Gateway certificate. Run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Secret
    type: kubernetes.io/tls
    metadata:
        name: istio-ingressgateway-certs
        namespace: istio-system
    data:
        tls.crt: $(echo "$KYMA_TLS_CERT" | base64)
        tls.key: $(echo "$KYMA_TLS_KEY" | base64)
    EOF
    ```
 
3. Update the `kyma-system` Namespace certificate:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
        name: ingress-tls-cert
        namespace: kyma-system
    data:
        tls.crt: $(echo "$KYMA_TLS_CERT" | base64)
    EOF
    ```
    
4. Update the `kyma-integration` Namespace certificate:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
        name: ingress-tls-cert
        namespace: kyma-integration
    data:
        tls.crt: $(echo "$KYMA_TLS_CERT" | base64)
    EOF
    ```

5. Restart the Ingress Gateway Pod to apply the new certificate:

    ```bash
    kubectl delete pod -l app=istio-ingressgateway -n istio-system
    ```
    
6. Restart the Pods in the `kyma-system` Namespace to apply the new certificate:

    ```bash
    kubectl delete pod -l tlsSecret=ingress-tls-cert -n kyma-system
    ```
    
7. Restart the Pods in the `kyma-integration` Namespace to apply the new certificate:

    ```bash
    kubectl delete pod -l tlsSecret=ingress-tls-cert -n kyma-integration
    ```
