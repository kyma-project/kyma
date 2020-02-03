---
title: Troubleshooting Knative Eventing upgrade
type: Troubleshooting
---

During the upgrade process, you can occasionally come across situations where the `kyma-installer` gets stuck at upgrading the `knative-eventing` component. This can happen because the upgrade process is unable to clear existing Knative Subscriptions created internally by Kyma's `event-bus-subscription-controller`. 
To verify this, look for the existing Knative Subscription resources present in the `kyma-system` Namespace:

```bash
    kubectl get subscriptions.eventing.knative.dev -n kyma-system
``` 

If you see some Knative subscription resources, this is probably the cause for the upgrade process of the `knative-eventing` Helm chart getting stuck. To solve this problem, edit the Knative Subscription:

```bash
    kubectl edit -n kyma-system subscriptions.eventing.knative.dev {NAME_OF_THE_KNATIVE_SUBSCRIPTION}
```
In the specification, look for the `finalizer` entry:
```yaml
    finalizers:
    - subscription.finalizers.kyma-project.io
```
If the finalizer is present, it means that Kubernetes was unable to clear this resource, therefore, blocking the upgrade process. To fix this, simply remove the finalizer and save your changes.

Removing the finalizer marks the resource for deletion and this should unblock your upgrade process.
