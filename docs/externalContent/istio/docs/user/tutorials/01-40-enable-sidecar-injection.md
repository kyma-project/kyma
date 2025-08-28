# Enable Istio Sidecar Proxy Injection

You can enable Istio sidecar proxy injection for either an entire namespace or a single Deployment. Learn how to perform both these operations.


## Enable Sidecar Injection for a Namespace

### Prerequisites
- You have the Istio module added.
- To use CLI instructions, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl). Alternatively, you can use Kyma dashboard.

### Context
Enabling Istio sidecar proxy injection for a namespace allows istiod to watch all Pod creation operations in this namespace and automatically inject newly created Pods with an Istio sidecar proxy.

>[!NOTE]
> A Pod is not injected with an Istio sidecar proxy if:
> - Istio sidecar proxy injection is disabled at the namespace level
> - The `sidecar.istio.io/inject` label on the Pod is set to `false`
> - The Pod's spec contains `hostNetwork: true`

### Procedure

<!-- tabs:start -->

#### Kyma Dashboard

1. Select the namespace where you want to enable sidecar proxy injection.
2. Choose **Edit**.
3. In the **UI Form** section, siwtch the toggle to enable Istio sidecar proxy injection.
4. Choose **Save**.

#### kubectl

Use the following command:

```bash
kubectl label namespace {YOUR_NAMESPACE} istio-injection=enabled
```

<!-- tabs:end -->

### Results
You've enabled Istio sidecar proxy injection for the specified namespace. The namespace is labeled with `istio-injection: enabled`, which means that all Pods created in it from now on have the Istio sidecar proxy injected.

## Enable Sidecar Injection for a Deployment

### Prerequisites
- You have the Istio module added.
- To use CLI instructions, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl). Alternatively, you can use Kyma dashboard.

### Context
Enabling Istio sidecar proxy injection for a Deployment injects an Istio sidecar proxy into all the Deployment's Pods.

>[!NOTE]
> A Pod is not injected with an Istio sidecar proxy if:
> - Istio sidecar proxy injection is disabled at the namespace level
> - The `sidecar.istio.io/inject` label on the Pod is set to `false`
> - The Pod's spec contains `hostNetwork: true`

### Procedure

<!-- tabs:start -->

#### Kyma Dashboard

1. Select the namespace of the Deployment for which you want to enable Istio sidecar proxy injection.
2. In the **Workloads** section, select **Deployments**.
3. Select the Deployment. 
4. Choose **Edit**.
5. In the **UI Form** section, switch the toggle to enable Istio sidecar proxy injection.
6. Choose **Save**.

#### kubectl

Run the following command:

```bash
kubectl patch -n {YOUR_NAMESPACE} deployments/{YOUR_DEPLOYMENT} -p '{"spec":{"template":{"metadata":{"labels":{"sidecar.istio.io/inject":"true"}}}}}'
```

<!-- tabs:end -->

### Results

You've enabled Istio sidecar proxy injection for the specified Deployment. The Deployment and all its Pods are labeled with `istio-injection: enabled`. All the Deployment's Pods are instantly injected with an Istio sidecar proxy.
