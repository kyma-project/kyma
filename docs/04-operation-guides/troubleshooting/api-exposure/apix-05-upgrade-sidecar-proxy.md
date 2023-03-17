---
title: Pods stuck in `Pending/Failed/Unknown` state after an upgrade
---

## Symptom

You cannot access services or Functions using the created APIRules. Kyma Gateway refuses the connection.
Some of your Pods are stuck in the `Pending/Failed/Unknown` state, or their sidecar proxy version differs from the installed Istio version.

## Cause

During the upgrade, Kyma triggers the rollout restart to the instance of resources to ensure full compatibility of the sidecar proxy with the newly installed Istio version. The only exceptions to this are standalone Pods (without owner reference) or Pods created by the Job controller. Pods that cannot complete the rollout restart process or the restart cannot be triggered because a Pod is a standalone resource or is created by the Job controller, get the `istio.reconciler.kyma-project.io/proxy-reset-warning` annotation with a brief explanation of the cause.

## Remedy

There are multiple reasons why Pods cannot become available, and each case must be troubleshot separately. After resolving the root cause, it's safe to perform a rollout restart manually and remove the annotation containing the proxy reset warning.

In the case of standalone Pods or Pods created by the Job controller, the owner should recreate Pods with incompatible sidecar proxy and verify if the Istio version matches the installed version. For more information, see the `istioctl version` and `istioctl proxy-status` commands.

