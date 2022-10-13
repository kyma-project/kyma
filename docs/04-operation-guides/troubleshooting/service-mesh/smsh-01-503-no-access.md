---
title: Can't access a Kyma endpoint (503 status code)
---

## Symptom

You try to access a Kyma endpoint and receive the `503` status code.

## Cause

This can be caused by a configuration error in the Istio Ingress Gateway. As a result, the endpoint you call is not exposed.

## Remedy

To fix this problem, restart the Pods of the Gateway.

1. List all available endpoints:

    ```bash
    kubectl get virtualservice --all-namespaces
    ```

2. Restart the Pods of the Istio Ingress Gateway to force them to recreate their configuration:

     ```bash
     kubectl delete pod -l app=istio-ingressgateway -n istio-system
     ```

If this solution doesn't work, you need to change the image of the Istio Ingress Gateway to allow further investigation. Kyma uses distroless Istio images which are more secure, but you cannot execute commands inside them. Follow this steps:

1. Edit the Istio Ingress Gateway Deployment:

    ```bash
    kubectl edit deployment -n istio-system istio-ingressgateway
    ```

2. Find the `istio-proxy` container and delete the `-distroless` suffix.

3. Check all ports used by the Istio Ingress Gateway:

    ```bash
    kubectl exec -ti -n istio-system $(kubectl get pod -l app=istio-ingressgateway -n istio-system -o name) -c istio-proxy -- netstat -lptnu
    ```

4. If ports `80` and `443` are not used, check the logs of the Istio Ingress Gateway container for errors related to certificates. Run:

    ```bash
    kubectl logs -n istio-system -l app=istio-ingressgateway -c ingress-sds
    ```

5. In the case of certificate-related issues, make sure that the `kyma-gateway-certs` and `kyma-gateway-certs-cacert` Secrets are available in the `istio-system` Namespace and that they contain proper data. Run:

    ```bash
    kubectl get secrets -n istio-system kyma-gateway-certs -oyaml
    kubectl get secrets -n istio-system kyma-gateway-certs-cacert -oyaml
    ```

<!-- Update step 6 once the long-lasting certificate is implemented. Probably, only the details about Gardener will be needed. -->
6. To regenerate a corrupted certificate, follow [this tutorial](../../../03-tutorials/00-security/sec-01-tls-certificates-security.md). If you are running Kyma provisioned through Gardener, follow [this troubleshooting guide](../security/sec-01-certificates-gardener.md) instead.

   >**NOTE**: Remember to switch back to the `distroless` image after you resolved the issue.
