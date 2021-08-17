---
title: Disable Istio sidecar proxy injection
---

By default, `istiod` watches all Pod creation operations on all Namespaces and injects the newly created Pods with a sidecar proxy.

You can disable sidecar proxy injection for either an entire Namespace or a single Deployment.

* To disable sidecar proxy injection for a Namespace, set the **istio-injection** label value to `disabled` for the Namespace in which you want to disable the sidecar proxy injection. Use this command:

   ```bash
   kubectl label namespace {YOUR_NAMESPACE} istio-injection=disabled
   ```

* To disable sidecar proxy injection for a Deployment, add this annotation to the Deployment configuration file: `sidecar.istio.io/inject: "false"`

Read the [Istio documentation](https://istio.io/docs/setup/kubernetes/additional-setup/sidecar-injection/) to learn more about sidecar proxy injection.

If there are issues with the Istio sidecar, you can check whether there is an [issue with the sidecar injection](../troubleshooting/smsh-03-istio-no-sidecar.md) or a [mismatching Istio version](../troubleshooting/smsh-04-istio-sidecar-version.md).
