---
title: Local Kyma deployment fails with pending Pods
---

## Symptom

Local Kyma deployment with k3d fails with one or more Pods in pending state.
Describing the pod reveals an error message like:

`0/2 nodes are available: 2 node(s) had taint {node.kubernetes.io/disk-pressure: }, that the pod didn't tolerate.`

## Cause

The underlying Docker environment ran out of resources (memory/cpu/disk). 
Thus, k3d marked all Kubernetes nodes with a taint "disk-pressure".

## Remedy

Verify the cause:

1. Find out which Pods are pending.
2. For the pending Pods, verify which error message you get by _describing_ the related deployment:
   For example, if the Istiod Pod is pending, run `kubectl -n istio-system describe deployment istiod` and `kubectl -n istio-system describe pod istiod-{POD_NAME}`.

Fix the issue:

1. In the preference settings of your Docker UI, adjust the Docker resource assignment.
2. Check whether there are sufficient resources (memory/cpu/disk) on the device.
3. Adjust the resources as needed. If you're not sure, use the default evaluation profile values, which you find in `profile-evaluation.yaml` of the respective component in the [`resources`](https://github.com/kyma-project/kyma/tree/main/resources) directory.
