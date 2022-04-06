---
title: Pods stuck in Pending/Failed/Unknown state after upgrade
---

## Symptom

You cannot access services or functions using the API Rules created. The Kyma Gateway refuses the connection.
Some of your Pods are stuck in the Pending/Failed/Unknown state, or it's sidecar proxy version differ from installed Istio version.

## Cause

During the upgrade Kyma will trigger rollout restart to instance of resources to ensure full compatibility of sidecar proxy with newly installed Istio version. Exception to this are standalone pods (without OwnerReference) or pods created by Job controller. Pods that cannot complete rollout restart process or rollout restart cannot be triggered because a pod is standalone or created by Job resource will get annotation `istio.reconciler.kyma-project.io/proxy-reset-warning` with brief explanation of the cause.

## Remedy

There are multiple reasons why Pods cannot become available and each case should be troubleshooted separately. After resolving root-cause it's safe to perform rollout restart manually and remove annotation with proxy reset warning.

In case of standalone pods or pods created by Job controller owner should recreate pods with incompatible sidecar proxy and verify if version match installed version (see commands: `istioctl version` and `istioctl proxy-status`)

