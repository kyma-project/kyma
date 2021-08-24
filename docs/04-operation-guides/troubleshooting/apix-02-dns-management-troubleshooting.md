---
title: External DNS managagement - troubleshooting 
---

See the list of possible issues related to the External DNS Management comopnent.

## "Connection refused" error

### Symptom

Even though the DNS Manaement setup is finished, you receive the `connection refused` error when you try to expose a service. It occurs when you call the service endpoint by sending a GET request. The error looks as follows:

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

On a non-Gardener clutser, and the DNSProvider or DNSEntry CR you cretaed is ignored by the controller.

### Cause

The following annotation was added to the CR.

```txt
 annotations:
     dns.gardener.cloud/class: garden
```

### Remedy

Remove the **metadata.annotations.dns.gardener.cloud/class** parameter from the CR.