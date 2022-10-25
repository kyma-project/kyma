---
title: Pods stuck in `Pending/Failed/Unknown` state after upgrade
---

## Symptom

You cannot access services or Functions using the APIRules created. The Kyma Gateway refuses the connection.
Some of your Pods are stuck in the Pending/Failed/Unknown state, or their sidecar proxy version differ from the installed Istio version.

## Cause

During the upgrade, Kyma triggers rollout restart to the instance of resources, in order to ensure full compatibility of sidecar proxy with the newly installed Istio version. Exception to this are standalone Pods (without owner reference) or Pods created by the Job controller. Pods that cannot complete the rollout restart process, or the restart cannot be triggered because a Pod is a standalone resource or is created by the Job controller, get annotation `istio.reconciler.kyma-project.io/proxy-reset-warning` with brief explanation of the cause.

## Remedy

There are multiple reasons why Pods cannot become available and each case must be troubleshot separately. After resolving the root cause it's safe to perform rollout restart manually and remove the annotation with proxy reset warning.

In case of standalone Pods or Pods created by the Job controller the owner should recreate Pods with incompatible sidecar proxy and verify if the Istio version match the installed version (see commands: `istioctl version` and `istioctl proxy-status`)

