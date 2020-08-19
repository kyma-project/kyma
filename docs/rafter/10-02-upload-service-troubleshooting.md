---
title: 'Upload Service returns "502 Bad Gateway"' 
type: Troubleshooting
---

It may happen that the Upload Service returns `502 Bad Gateway` with the `The specified bucket does not exist` message. It means that the bucket has been removed or renamed.

If the bucket is removed, delete the ConfigMap:

```bash
kubectl -n kyma-system delete configmaps rafter-upload-service
```

If the bucket exists but with a different name, set a proper name in the ConfigMap:

```bash
kubectl -n kyma-system edit configmaps rafter-upload-service
```

After that, restart the Upload Service - it will create a new bucket or use the renamed one:

```bash
kubectl -n kyma-system delete pod -l app.kubernetes.io/name=upload-service
```