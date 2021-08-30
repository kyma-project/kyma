---
title: External DNS managagement - Resource ignored by the controller
---

## Symptom

If you use a custom domain to expose a service and you have a non-Gardener clutser, the DNSProvider or DNSEntry CR you cretaed could be ignored by the controller.

## Cause

The following annotation was added to the CR(s).

```txt
 annotations:
     dns.gardener.cloud/class: garden
```

## Remedy

Remove the **metadata.annotations.dns.gardener.cloud/class** parameter from the DNSProvider and/or DNSEntry CR.
