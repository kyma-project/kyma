---
title: Sidecar Proxy Injection
type: Details
---

By default, the Istio sidecar injector watches all Pod creation operations on all Namespaces and injects the newly created Pods with a sidecar proxy.

You can disable sidecar proxy injection for either an entire Namespace or a single Deployment.

* To disable sidecar proxy injection for a Namespace, set the **istio-injection** label value to `disabled` for the Namespace in which you want to disable the sidecar proxy injection. Use this command: `kubectl label namespace {YOUR_NAMESPACE} istio-injection=disabled`                                                                                                                                                                                
* To disable sidecar proxy injection for a Deployment, add this annotation to the Deployment configuration file: `sidecar.istio.io/inject: "false"`

Read the [Installing the Sidecar](https://istio.io/docs/setup/kubernetes/additional-setup/sidecar-injection/) document to learn more about sidecar proxy injection.
