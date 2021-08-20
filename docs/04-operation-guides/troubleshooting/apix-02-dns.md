---
title: Connection refused
---

## Symptom

The following `connection refused` error occurs when calling the service endpoint by sending a GET request:

```txt
curl: (7) Failed to connect to httpbin.kyma-goat.ga port 443: Connection refused
```

## Cause

Incorrect IP provided.

## Remedy

Check if the IP address provided as the value of the **spec.targets** parameter of the DNS Entry Custom Resource (CR).
