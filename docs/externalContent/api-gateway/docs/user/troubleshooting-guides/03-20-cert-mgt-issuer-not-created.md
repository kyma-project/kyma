# Certificate Management - Issuer Not Created

## Symptom

When you try to create an Issuer custom resource (CR) using `cert.gardener.cloud/v1alpha1`, the resource is not created. There are no logs in the `cert-management` controller.

## Cause

The namespace in which the Issuer CR was created is incorrect. By default, `cert-management` watches the `default` namespace for all Issuer CRs.

## Solution

Make sure that you created the Issuer CR in the `default` namespace. Run:

```bash
kubectl get issuers -A
```

If you want to create the Issuer CR in a different namespace, adjust the `cert-management` settings during the installation.
