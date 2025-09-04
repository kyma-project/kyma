# Istio Sidecar Proxy Injection Issues

## Symptom

Some Pods don't have an Istio sidecar proxy injected.

## Cause

By default, the Istio module does not automatically inject an Istio sidecar proxy into any Pods you create. To inject a Pod with an Istio sidecar proxy, you must explicitly enable injection for the Pod's Deployment or for the entire namespace. If you have done this and the sidecar is still not installed, follow the remedy steps to identify which settings are preventing the injection.

A Pod is not injected with an Istio sidecar proxy if:
- Istio sidecar proxy injection is disabled at the namespace level
- The **sidecar.istio.io/inject** label on the Pod is set to `false`
- The Pod's `spec` contains `hostNetwork: true`

## Solution

Find out which Pods do not have Istio sidecar proxy injection enabled and why. You can either inspect Pods across all namespaces, focus on a specific namespace, or verify why a selective Pod is not injected.

### Check Pods in All Namespaces

1. Download the [script](../../assets/sidecar-analysis.sh).
2. Run the script.

    ```bash
    ./sidecar-analysis.sh
    ```

3. You get an output similar to this one:

    ```bash
    See all Pods that are not part of the Istio service mesh:
    Pods in namespaces labeled with "istio-injection=disabled":
        - namespace/Pod
        ...
    Pods labeled with "sidecar.istio.io/inject=false" in namespaces labeled with "istio-injection=enabled":
        - namespace/Pod
        ...
    Pods not labeled with "sidecar.istio.io/inject=true" in unlabeled namespaces:
        - namespace/Pod
        ...
    ```
4. To learn how to include a Pod into the Istio service mesh, see [Enabling Istio Sidecar Proxy Injection](../tutorials/01-40-enable-sidecar-injection.md).

### Check Pods in a Selective Namespace

1. Download the [script](../../assets/sidecar-analysis.sh).
2. Run the script passing the namespace name as a parameter.

    ```bash
    ./sidecar-analysis.sh {NAMESPACE}
    ```

3. You get an output similar to this one:

    ```bash
    In the namespace default, the following Pods are not part of the Istio service mesh:
     - Pod
     - Pod
     ...
    Reason: The namespace default has Istio sidecar proxy injection disabled, so none of its Pods have been injected with an Istio sidecar proxy.
    ```
4. To learn how to include a Pod into the Istio service mesh, see [Enabling Istio Sidecar Proxy Injection](../tutorials/01-40-enable-sidecar-injection.md).

### Check a Selective Pod

<!-- tabs:start -->

#### Kyma Dashboard

1. Go to the Pod's namespace.
2. Check if Istio sidecar proxy injection is enabled at the namespace level.
    Verify if the `Labels` section contains `istio-injection=disabled` or `istio-injection=enabled`. If Istio sidecar proxy injection is disabled at the namespace level, none of its Pods are injected with an Istio sidecar proxy.
3. Check if Istio sidecar proxy injection is enabled for the Pod's Deployment.
   1. In the **Workloads** section, choose **Deployments**
   2. Choose **Edit**. 
   3. In the **UI Form** section, check if the `Enable Sidecar Injection` toggle is switched.
4. Check if the label `sidecar.istio.io/inject: false` is set on a Pod.
   1. In the **Workloads** section, choose **Pods**.
   2. Search for `sidecar.istio.io/inject: false`. 
   If your Pod is displayed on the list, it has the label set.

#### kubectl

1. To check if Istio sidecar proxy injection is enabled at the namespace level, run the command:

    ```bash
    kubectl get namespaces {POD_NAMESPACE} -o jsonpath='{ .metadata.labels.istio-injection }'
    ```

2. To check if Istio sidecar proxy injection is enabled for the Pod's Deployment, run the command:

    ```bash
    kubectl get deployments {POD_DEPLOYMENT} -n {NAMESPACE} -o jsonpath='{ .spec.template.metadata.labels }'
    ```
3. Check if the label `sidecar.istio.io/inject: false` is set on a Pod:
    ```bash
    kubectl get pod {POD} -n default -o=jsonpath='{.metadata.labels.sidecar\.istio\.io/inject}
    ```
4. To learn how to include a Pod into the Istio service mesh, see [Enabling Istio Sidecar Proxy Injection](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/enabling-istio-sidecar-proxy?locale=en-US&state=DRAFT&version=Internal).

<!-- tabs:end -->
