---
title: Local Kyma deployment fails with pending Pods
---

## Symptom

Local Kyma deployment with k3d fails with one or more Pods in pending state.
Describing the pod reveals an error message like `0/2 nodes are available: 2 node(s) had taint {node.kubernetes.io/disk-pressure: }, that the pod didn't tolerate.`

## Cause

K3D marked all kubernetes nodes with a taint "disk-pressure" as the underlying Docker environment run out of resources (memory/cpu/disk).

## Remedy

1. Verify which Pods are pending.
2. For the pending Pods, verify which error message you get by _describing_ the related deployment:
   Run `kubectl -n istio-system describe deployment istiod` and `kubectl -n istio-system describe pod istiod-{POD_NAME}`.
3. In the preference settings of your Docker UI, adjust the Docker resource assignment.
4. Check if whether there are sufficient resources (memory/cpu/disk) on the device.
5. Adjust the resources as needed.
