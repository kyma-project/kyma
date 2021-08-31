---
title: Certificate management - Issuer not created
---

## Symptom

Whey you try to create an Issuer CR using `cert.gardener.cloud/v1alpha1`, the resource is no created. There are no logs in the `cert-management` controller.

## Cause

By default, the `cert-management` watches the `default` Namespace for all Issuer CRs.

## Remedy

Make sure if you created the Issuer CR in the `default` Namespace. Run:

```bash
kubectl get issuers -A
```

If you want to create the Issuer CR in a different Namespace, adjust the `cert-management` settings during the installation.
