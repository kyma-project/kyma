# Resource Ignored by the Controller

## Symptom

If you have a non-Gardener cluster, the DNSProvider and/or DNSEntry custom resource (CR) you created are ignored by the controller.

## Cause

The following annotation was added to the CR(s).

```txt
 annotations:
     dns.gardener.cloud/class: garden
```

## Solution

Remove the **metadata.annotations.dns.gardener.cloud/class** parameter from the DNSProvider and/or DNSEntry CR.
