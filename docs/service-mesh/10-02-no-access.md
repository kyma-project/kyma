---
title: Can't access Console UI or other endpoints
type: Troubleshooting
---

The `503` status code received when you try to access the Console UI or any other endpoint in Kyma can be caused by a configuration error in the Istio Ingress Gateway. As a result, the endpoint you call is not exposed.
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
    kubectl scale --replicas 0 -n istio-system deploy/istio-ingressgateway
    kubectl edit deployment -n istio-system istio-ingressgateway
    kubectl scale --replicas 1 -n istio-system deploy/istio-ingressgateway
    ```
   
2. Find the `istio-proxy` container and delete the `-distroless` suffix.

3. Check all ports used by the Istio Ingress Gateway:

    ```bash
    kubectl exec -ti -n istio-system $(kubectl get pod -l app=istio-ingressgateway -n istio-system -o name) -c istio-proxy -- netstat -lptnu
    ```

4. If ports `80` and `443` are not used, restart the Pods of the Istio Ingress Gateway to force them to recreate their configuration:

    ```bash
    kubectl delete pod -l app=istio-ingressgateway -n istio-system
    ```
