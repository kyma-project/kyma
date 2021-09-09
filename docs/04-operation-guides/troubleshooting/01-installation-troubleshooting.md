---
title: Installation troubleshooting
type: Troubleshooting
---

## Component doesn't work after successful installation

If the installation is successful but a component does not behave in the expected way, see if all deployed Pods are running. Run this command:

```bash
kubectl get pods --all-namespaces
```

The command retrieves all Pods from all Namespaces, the status of the Pods, and their instance numbers. Check if the status is `Running` for all Pods. If any of the Pods that you require do not start successfully, install Kyma again.
