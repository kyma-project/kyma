---
title: External DNS management - could not resolve host
---

## Symptom

After all the steps required to [prepare your custom domain](../../../03-tutorials/00-api-exposure/apix-01-own-domain.md) are finished, you receive the `could not resolve host` error when you try to expose a service. It shows up when you call the service endpoint by sending a GET request. The error looks as follows:

```txt
curl: (6) Could not resolve host: httpbin.kyma-goat.ga
```

## Cause

The error could result from:

- Timing issues during the DNSEntry creation
- VPN connection on - issues related to DNS servers managed by your VPN provider
- Invalid DNS settings on your OS

## Remedy

- Wait for the DNSEntry custom resource to be created. Check if it's in the `Ready` status with the following command:

```bash
kubectl get dnsentry.dns.gardener.cloud dns-entry
```

- Turn the VPN off.

- Log in to your DNS provider's console and check if your new domain entry was added.

- Check if your local DNS configuration in `/etc/hosts`, or an equivalent file on your OS, contains an entry for the target host. If it does, remove the entry.
