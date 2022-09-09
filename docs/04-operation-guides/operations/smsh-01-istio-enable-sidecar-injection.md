---
title: Enable automatic Istio sidecar proxy injection
---

Enabling automatic sidecar injection allows `istiod` to watch all Pod creation operations on all Namespaces, which should be part of Istio Service Mesh, and inject the newly created Pods with a sidecar proxy.

You can enable sidecar proxy injection for either an entire Namespace or a single Deployment.

* To enable sidecar proxy injection for a Namespace, set the **istio-injection** label value to `enabled` for the Namespace in which you want to enable the sidecar proxy injection. Use this command:

   ```bash
   kubectl label namespace {YOUR_NAMESPACE} istio-injection=enabled
   ```

* To enable sidecar proxy injection for a Deployment, add this to the Deployment configuration file as either a label or an annotation: `sidecar.istio.io/inject: "true"`

Note that the Namespace label takes precedence over the Pod label or annotation.

Read the [Istio documentation](https://istio.io/docs/setup/kubernetes/additional-setup/sidecar-injection/) to learn more about sidecar proxy injection and consider [benefits of having the sidecar container inside your application pod](../../01-overview/main-areas/service-mesh/smsh-03-istio-sidecars-in-kyma.md).

If there are issues with the Istio sidecar, you can check whether there is an [issue with the sidecar injection](../troubleshooting/service-mesh/smsh-03-istio-no-sidecar.md) or a [mismatching Istio version](../troubleshooting/service-mesh/smsh-04-istio-sidecar-version.md).