# Secret with certificates is missing

## Symptom

Secret `kyma-gateway-certs` is not found,

```bash
kubectl get -n istio-system secrets
```

## Cause

This can only be caused by a mistake, for example accidental removal of that secret.

## Remedy

The certificate will be restored automatically in next reconciliation loop.
