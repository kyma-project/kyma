# Could Not Resolve Host

## Symptom

After you have completed all the steps required to [set up your custom domain](../../tutorials/01-10-setup-custom-domain-for-workload.md), you receive the `could not resolve host` error when you try to expose a Service. It shows up when you call the Service endpoint by sending a `GET` request. The error looks as follows:

```txt
curl: (6) Could not resolve host: httpbin.mydomain.com
```

## Cause

The error could result from:

- Timing issues during the DNSEntry creation
- VPN connection on - issues related to DNS servers managed by your VPN provider
- Invalid DNS settings on your OS

## Solution

- Wait for the DNSEntry custom resource to be created. Check if it has the `Ready` status using the following command:

    ```bash
    kubectl get dnsentry.dns.gardener.cloud dns-entry
    ```

- Turn the VPN off.

- Log in to your DNS provider's console and check if your new domain entry was added.

- Check if your local DNS configuration in `/etc/hosts`, or an equivalent file on your OS, contains an entry for the target host. If it does, remove the entry.
