---
title: External DNS management - connection refused
---

## Symptom

If you use a custom domain, you receive the `connection refused` error when you try to expose a service. It shows up when you call the service endpoint by sending a GET request. The error looks as follows:

```txt
curl: (7) Failed to connect to httpbin.kyma-goat.ga port 443: Connection refused
```

## Cause

DNS resolves to an incorrect IP address.

## Remedy

Check if the IP address provided as the value of the **spec.targets** parameter of the DNS Entry Custom Resource (CR) is the IP address of the Ingress Gateway you are using. To check the Ingress Gateway IP, run:

```bash
kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`
```
