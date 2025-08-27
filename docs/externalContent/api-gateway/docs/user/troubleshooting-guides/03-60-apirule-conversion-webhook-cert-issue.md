# APIRule Conversion Webhook Not Working Because of Certificate Issue

## Symptoms

- There is an error related to the conversion webhook certificate when fetching or creating an `APIRule` resource, for example:
  ```bash
  Error from server: conversion webhook for gateway.kyma-project.io/v1beta1, Kind=APIRule failed: Post "https://api-gateway-webhook-service.kyma-system.svc:9443/convert?timeout=30s": x509: certificate has expired or is not yet valid


- There is an error in the `api-gateway-controller-manager` logs with the message `"Secret \"api-gateway-webhook-certificate\" not found"`, for example:    
  ```bash
  ERROR	Reconciler error	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"api-gateway-webhook-certificate","namespace":"kyma-system"}, "namespace": "kyma-system", "name": "api-gateway-webhook-certificate", "reconcileID": "a808b99f-6db6-47f5-a82e-8176811238ac", "error": "Secret \"api-gateway-webhook-certificate\" not found"}

## Cause

The Secret `api-gateway-webhook-certificate` was deleted, or the automatic certificate rotation did not work.

## Solution

Restart `api-gateway-controller-manager` to recreate the Secret:

```bash
kubectl rollout restart deployment -n kyma-system api-gateway-controller-manager
```
