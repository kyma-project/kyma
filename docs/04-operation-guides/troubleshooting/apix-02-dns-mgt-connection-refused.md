---
title: External DNS managagement - Connection refused
---

## Symptom

If you use a custom domain, you could receive the `connection refused` error when you try to expose a service. It shows up when you call the service endpoint by sending a GET request. The error looks as follows:

```txt
curl: (7) Failed to connect to httpbin.kyma-goat.ga port 443: Connection refused
```

## Cause

Incorrect IP provided.

## Remedy

Check if the IP address provided as the value of the **spec.targets** parameter of the DNS Entry Custom Resource (CR) is correct.
