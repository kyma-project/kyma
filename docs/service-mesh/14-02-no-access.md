---
title: Can't access Console UI or other endpoints
type: Troubleshooting
---

The `503` status code received when you try to access the Console UI or any other endpoint in Kyma can be caused by a configuration error in the Istio Ingress Gateway. As a result, the endpoint you call is not exposed.
To fix this problem, restart the Pods of the Gateway.

1. List all available endpoints:
  ```
  kubectl get virtualservice --all-namespaces
  ```

2. Check all ports used by the Istio Ingress Gateway:
  ```
  kubectl exec -t -n istio-system $(kubectl get pod -l app=istio-ingressgateway -n istio-system | grep "istio-ingressgateway" | awk '{print $1}') -- netstat -lptnu
  ```

3. If ports `80` and `443` are not used, restart the Pods of the Istio Ingress Gateway to force them to recreate their configuration:
  ```
  kubectl delete pod -l app=istio-ingressgateway -n istio-system
  ```
