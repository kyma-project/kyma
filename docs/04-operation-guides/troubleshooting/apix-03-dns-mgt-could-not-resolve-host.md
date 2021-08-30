---
title: External DNS managagement - Could not resolve host
---

## Symptom

If you use a custom domain, you could receive the `could not resolve host` error when you try to expose a service. It shows up when you call the service endpoint by sending a GET request. The error looks as follows:

```txt
curl: (6) Could not resolve host: httpbin.kyma-goat.ga
```

## Cause

The error could result from:

- Timing issues during the DNS Entry creation
- VPN connection
- `etc/host` or `etc/resolve-conf` settings

## Remedy

- Wait for the DNS Entry to be created. To check the CR status, run:

```bash
kubectl get dnsentry.dns.gardener.cloud dns-entry
```

- Turn the VPN off.

- Check if the subdomain is not added to the `etc/host` file or check the `etc/resolve-conf` settings.
