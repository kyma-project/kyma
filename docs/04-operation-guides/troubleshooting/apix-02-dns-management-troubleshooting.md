---
title: DNS managagement - troubleshooting 
---


## `"Connection refused" error`

### Symptom

When you use your own domain to expose a service, you may receive the `connection refused` error. It occurs when you call the service endpoint by sending a GET request. The error looks as follows:

```txt
curl: (7) Failed to connect to httpbin.kyma-goat.ga port 443: Connection refused
```

### Cause

Incorrect IP provided.

### Remedy

Check if the IP address provided as the value of the **spec.targets** parameter of the DNS Entry Custom Resource (CR) is correct.

## "Could not resolve host" error
---

### Symptom

```txt
curl: (6) Could not resolve host: httpbin.kyma-goat.ga
```

### Cause

### Remedy

## Resource ignored by the controller
---

### Symptom

```txt
curl: (6) Could not resolve host: httpbin.kyma-goat.ga
```

### Cause

### Remedy

