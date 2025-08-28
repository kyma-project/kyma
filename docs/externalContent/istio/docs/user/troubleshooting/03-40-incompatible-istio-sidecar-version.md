<!-- open-source-only -->
# Incompatible Istio Sidecar Version After the Istio Module's Update
## Symptom

After the Istio module's update, the Istio custom resource (CR) is in the `Warning` state, and mesh connectivity is disrupted. When you click on the warning in Kyma dashboard or run `kubectl get istio default -n kyma-system -o jsonpath='{.status.description}'`, you get the following message: 

```
Some Pods with Istio sidecar injection failed to restart. To learn more about the warning, see kyma-system/istio-controller-manager logs: Istio Controller could not restart one or more Istio-injected Pods.`
```

## Cause

The sidecar version in Pods must match the installed Istio version to ensure proper mesh connectivity. During an upgrade of the Istio module to a new version, Istio Operator's `ProxySidecarReconcilation` component performs a rollout for most common workload types ensuring that the injected Istio sidecar proxies are updated correctly.
However, if a resource is a Job, a ReplicaSet that is not managed by any Deployment, or a Pod that is not managed by any other resource, the restart cannot be performed automatically.

You must manually restart such workloads to ensure proper functionality with the updated Istio version.

## Solution

<!-- tabs:start -->

#### Kyma dashboard

1. Choose **Modify Modules**.
2. Select the Istio module.
   The `ProxySidecarRestartSucceeded` reconciliation condition has the status `False` and the reason: `ProxySidecarManualRestartRequired`. The message contains the list of workloads that you must restart manually, for example:
   ```
   The sidecars of the following workloads could not be restarted: test/httpbin
   ```
3. Restart the listed workloads so that new Istio sidecars are injected into the Pods.

#### kubectl

1. To get the Istio CR, run:
   ```
   kubectl get istio default -n kyma-system -o yaml
   ```

2. To learn if any Pods or workloads require a manual restart, see the message in the **status.conditions** section of the Istio CR, for example:
   ```yaml
   - lastTransitionTime: "2025-01-23T10:02:08Z"
     message: 'The sidecars of the following workloads could not be restarted: test/httpbin'
     reason: ProxySidecarManualRestartRequired
     status: "False"
     type: ProxySidecarRestartSucceeded
   ```
3. Restart the listed workloads so that new Istio sidecars are injected into the Pods.

<!-- tabs:end -->
