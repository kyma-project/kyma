---
title: Troubleshooting Knative Eventing upgrade
type: Troubleshooting
---

During the upgrade process, you can occasionally come across situations where the `kyma-installer` gets stuck at upgrading `knative-eventing`. This can happen because the upgrade process was unable to clear existing Knative Subscriptions, created internally by Kyma's `event-bus-subscription-controller`. 
To verify this, look for the existing Knative Subscription resources present in the `kyma-system` Namespace:

```bash
    kubectl get subscriptions.eventing.knative.dev -n kyma-system
``` 

if you see some knative subscription resources, this can probably be the cause of the never ending upgrade process of the `knative-eventing` helm chart. To rectify this problem, edit the knative subscription by running the following command.

```bash
    kubectl edit -n kyma-system subscriptions.eventing.knative.dev <NAME_OF_THE_KNATIVE_SUBSCRIPTION>
```
This opens the Custom Resource in your text editor, and you should be able to edit the current specifications. Now, verify if there's a `finalizer` entry in the specification, as below:
```yaml
    finalizers:
    - subscription.finalizers.kyma-project.io
```
This would imply that kubernetes was unable to clear this resource and this state, consequently, blocks the upgrade process. Hence, you should simply remove the above finalizer and save your changes (`:wq` command if `vim` is your default text editor)

Removing the finalizer would mark the resource for deletion and this should unblock your upgrade process.
