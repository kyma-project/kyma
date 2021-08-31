---
title: External DNS management - resource ignored by the controller
---

## Symptom

If you have a non-Gardener clutser, the DNSProvider or DNSEntry CR you cretaed are ignored by the controller.

## Cause

The following annotation was added to the CR(s).

```txt
 annotations:
     dns.gardener.cloud/class: garden
```

## Remedy

Remove the **metadata.annotations.dns.gardener.cloud/class** parameter from the DNSProvider and/or DNSEntry CR.
