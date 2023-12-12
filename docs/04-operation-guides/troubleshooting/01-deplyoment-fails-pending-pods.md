---
title: Local Kyma Deployment Fails With Pending Pods
---

## Symptom

Local Kyma deployment with k3d fails with one or more Pods in state `Pending`.
Describing the Pod reveals such an error message:

```bash
0/2 nodes are available: 2 node(s) had taint {node.kubernetes.io/disk-pressure: }, that the pod didn't tolerate.
```

## Cause

The underlying Docker environment ran out of resources (memory/CPU/disk). 
Thus, k3d marked all Kubernetes nodes with a taint `disk-pressure`.

## Remedy

Verify the cause:

1. Find out which Pods are pending:
   ```bash
   kubectl --all-namespaces get pods
   ```
2. For the pending Pods, verify which error message you get:
   ```bash
   kubectl -n {POD_NAMESPACE} describe pod {POD_NAME}
   ```

Fix the issue:

1. In the preferences of your Docker Desktop, adjust the Docker resource assignment.
2. Check whether there are sufficient resources (memory/CPU/disk) on the device.
3. Adjust the resources as needed. If you're not sure, try 8 GB for the evaluation profile.
