---
title: Upload Service returns 502 code - Bad Gateway
type: Troubleshooting
---

It may happen that the Upload Service returns 502 Bad Gateway with `The specified bucket does not exist` message. It means that the bucket has been removed or renamed.

If the bucket is removed, then delete ConfigMap:

```bash
kubectl -n kyma-system delete configmaps rafter-upload-service
```

If bucket exists, but with a different name, then set proper name in ConfigMap:

```bash
kubectl -n kyma-system edit configmaps rafter-upload-service
```

After that restart Upload Service - it will create a new bucket or use the renamed one:

```bash
kubectl -n kyma-system delete pod -l app.kubernetes.io/name=upload-service
```