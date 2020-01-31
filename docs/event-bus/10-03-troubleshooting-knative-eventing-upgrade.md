---
title: Troubleshooting Knative Eventing Upgrade
type: Troubleshooting
---

During the upgrade process, there might be certain sporadic instances where the `kyma-installer` gets stuck at the upgrade process of `knative-eventing`. This can happen due to the reason that the upgrade process was unable to clear existing knative subscriptions which are created internally by Kyma's `event-bus-subscription-controller`. 
To verify this check for the existing knative Subscription resources present in the `kyma-system` namespace. In order to do so, execute the following statement.

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
