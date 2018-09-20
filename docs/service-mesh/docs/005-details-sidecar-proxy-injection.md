---
title: Sidecar Proxy Injection
type: Details
---

By default, the Istio Sidecar Injector watches all Pod creation operations on all Namespaces but it does not inject the newly created Pods with a sidecar proxy.

To enable the sidecar proxy injection, set the **istio-injection** label value to `enabled` for the Namespace in which you want to enable the sidecar proxy injection. Use this command:

```
kubectl label namespace {YOUR_NAMESPACE} istio-injection=enabled
```

With the sidecar proxy injection enabled, you can inject the sidecar to Pods of a selected deployment in the given Namespace. Add this annotation to the deployment configuration file:
```
sidecar.istio.io/inject: "true"
```

Read the [Installing the Istio Sidecar](https://istio.io/docs/setup/kubernetes/sidecar-injection.html) document to learn more about sidecar proxy injection.
