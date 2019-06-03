---
title: Can't access Console UI or other endpoints
type: Troubleshooting
---

If you get a code `503` error when you try to access the Console UI or any other endpoint in Kyma, that means the Istio Ingress gateway failed to expose them. This problem is easily fixed by restarting the Pods of the gateway.

List all of the available endpoints:
```
kubectl get virtualservice --all-namespaces
```

Check all of the ports used by the gateway:
```
kubectl exec -t -n istio-system $(kubectl get pod -l app=istio-ingressgateway -n istio-system | grep "istio-ingressgateway" | awk '{print $1}') -- netstat -lptnu
```

If ports `80` and `443` are not used, restart the Pods of the gateway to force them to recreate their configuration:
```
kubectl delete pod -l app=istio-ingressgateway -n istio-system
```
