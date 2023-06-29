---
title: Enable automatic Istio sidecar proxy injection
---

Enabling automatic sidecar injection allows `istiod` to watch all Pod creation operations on all Namespaces, which should be part of Istio Service Mesh, and inject the newly created Pods with a sidecar proxy.

You can enable sidecar proxy injection for either an entire Namespace or a single Deployment.

* To enable sidecar proxy injection for a Namespace, set the **istio-injection** label value to `enabled` for the Namespace in which you want to enable the sidecar proxy injection. Use this command:

   ```bash
   kubectl label namespace {YOUR_NAMESPACE} istio-injection=enabled
   ```

* To enable sidecar proxy injection for a Deployment, add this label to the Deployment configuration file: `sidecar.istio.io/inject: "true"`

Note that if the sidecar proxy injection is disabled at the Namespace level or the `sidecar.istio.io/inject` label on a Pod is set to `false`, the sidecar proxy is not injected.

Read the [Istio documentation](https://istio.io/docs/setup/kubernetes/additional-setup/sidecar-injection/) to learn more about sidecar proxy injection and consider [benefits of having the sidecar container inside your application pod](../../01-overview/service-mesh/smsh-03-istio-sidecars-in-kyma.md).

If there are issues with the Istio sidecar, you can check whether there is an [issue with the sidecar injection](../troubleshooting/service-mesh/smsh-03-istio-no-sidecar.md) or a [mismatching Istio version](../troubleshooting/service-mesh/smsh-04-istio-sidecar-version.md).

## Check whether your workloads have automatic Istio sidecar injection enabled

You can easily check whether your workloads have automatic Istio sidecar injection enabled by running [this script](../assets/sidecar-analysis.sh). You can either pass a **namespace** parameter to the script or run it with no parameter.

If no parameter is passed, the execution output will contain Pods from all Namespaces that don't have automatic Istio sidecar injection enabled, whereas passing the parameter results in the analysis of only the given  Namespace.

The script outputs the information in `{namespace}/{pod}` if run for all Namespaces and in `{pod}` form for a specific Namespace.

* Run the script

```bash
./sidecar-analysis.sh {namespace}
```

* Example output

  * `./sidecar-analysis.sh`

  ```
  Pods out of istio mesh:
    In namespace labeled with "istio-injection=disabled":
      - sidecar-disabled/some-pod
    In namespace labeled with "istio-injection=enabled" with pod labeled with "sidecar.istio.io/inject=false":
      - sidecar-enabled/some-pod
    In not labeled ns with pod not labeled with "sidecar.istio.io inject=true":
      - no-label/some-pod
  ```

  * `./sidecar-analysis.sh some-namespace`

  ```
  Pods out of istio mesh in namespace some-namespace:
    - some-pod
  ```
