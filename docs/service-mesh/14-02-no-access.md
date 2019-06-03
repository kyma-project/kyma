---
title: Can't access Console UI or other endpoints
type: Troubleshooting
---

A response of code `503` error when trying to access the Console UI or any other endpoint in Kyma can be caused by a configuration error in the Istio Ingress gateway. As a result, the endpoint you call is not exposed.
This problem is easily fixed by restarting the Pods of the gateway.

1. List all of the available endpoints:
  ```
  kubectl get virtualservice --all-namespaces
  ```

2. Check all of the ports used by the gateway:
  ```
  kubectl exec -t -n istio-system $(kubectl get pod -l app=istio-ingressgateway -n istio-system | grep "istio-ingressgateway" | awk '{print $1}') -- netstat -lptnu
  ```

3. If ports `80` and `443` are not used, restart the Pods of the gateway to force them to recreate their configuration:
  ```
  kubectl delete pod -l app=istio-ingressgateway -n istio-system
  ```
