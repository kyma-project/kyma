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
If needed, you can also trigger the reconciliation by restarting the `api-gateway-controller-manager` deployment:

```bash
kubectl -n kyma-system rollout restart deployment api-gateway-controller-manager
```
