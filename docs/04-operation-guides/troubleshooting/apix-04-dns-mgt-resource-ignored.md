---
title: External DNS management - resource ignored by the controller
---

## Symptom

If you have a non-Gardener cluster, the DNS Provider and/or DNS Entry CR you created are ignored by the controller.

## Cause

The following annotation was added to the CR(s).

```txt
 annotations:
     dns.gardener.cloud/class: garden
```

## Remedy

Remove the **metadata.annotations.dns.gardener.cloud/class** parameter from the DNS Provider and/or DNS Entry CR.
