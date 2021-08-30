---
title: External DNS managagement - troubleshooting 
---

See the list of possible issues related to the External DNS Management comopnent.

## "Connection refused" error

### Symptom

The `connection refused` error occurs when you try to expose a service. It shows up when you call the service endpoint by sending a GET request. The error looks as follows:

```txt
curl: (7) Failed to connect to httpbin.kyma-goat.ga port 443: Connection refused
```

### Cause

Incorrect IP provided.

### Remedy

Check if the IP address provided as the value of the **spec.targets** parameter of the DNS Entry Custom Resource (CR) is correct.

## "Could not resolve host" error

### Symptom

The `could not resolve host` error occurs when you try to expose a service. It shows up when you call the service endpoint by sending a GET request. The error looks as follows:

```txt
curl: (6) Could not resolve host: httpbin.kyma-goat.ga
```

### Cause

The error could result from:

- Timing issues during the DNS Entry creation
- VPN connection
- `etc/host` or `etc/resolve-conf` settings

### Remedy

- Wait for the DNS Entry to be created. To check the CR status, run:

```bash
kubectl get dnsentry.dns.gardener.cloud dns-entry
```

- Turn the VPN off.

- Check if the subdomain is not added to the `etc/host` file or check the `etc/resolve-conf` settings.

## Resource ignored by the controller

### Symptom

On a non-Gardener clutser, and the DNSProvider or DNSEntry CR you cretaed is ignored by the controller.

### Cause

The following annotation was added to the CR(s).

```txt
 annotations:
     dns.gardener.cloud/class: garden
```

### Remedy

Remove the **metadata.annotations.dns.gardener.cloud/class** parameter from the DNSProvider and/or DNSEntry CR.
