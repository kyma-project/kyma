<!-- open-source-only -->
# Changes to Istio Resources Are Not Reverted After Reconciliation

## Symptoms

- The Istio Ingress Gateway Deployment, Istiod Deployment, or  Istio CNI DaemonSets fails to start.
- Any of these resources contains a non-default container that is unable to start.
- The Istio custom resource is in the `Error` state.

## Cause

- If an external component, like a mutating webhook, adds a container to the Istio Ingress Gateway Deployment, Istiod Deployment, or Istio CNI DaemonSets, and that new container fails to start for any reason, the Pod is unable to reach the `Running` state. As a result, the Istio module reconciliation fails.

## Solution

1. Check if a given resource's Pod template has been modified to include an additional container.
2. Check if the newly injected container fails to start.

> [!TIP]
> To retrieve information for the above steps, you can use the command `kubectl describe {RESOURCE_NAME} -n istio-system -o yaml`.

3. If the container is unable to start, remove it from the Pod template of the given resource.

   `kubectl edit {RESOURCE_NAME} -n istio-system`

4. If the error persists, check if there is a mutating webhook in the cluster that is modifying Istio resources.