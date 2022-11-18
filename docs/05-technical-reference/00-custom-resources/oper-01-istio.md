---
title: Istio
type: Custom Resource
---

The `istios.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the Reconciler uses to configure and install Istio. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```shell
kubectl get crd istios.operator.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample Istio custom resource (CR) that the Reconciler uses to configure and install Istio. This example shows the single supported **numTrustedProxies** configuration setting. There must be only one Istio CR on the cluster.

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: Istio
metadata:
  name: istio
  labels:
    app.kubernetes.io/name: istio
spec:
  config:
    numTrustedProxies: 1
```

This table lists all the possible parameters of a given resource together with their descriptions:

| Field   |      Mandatory      |  Description |
|---|:---:|---|
| **metadata.name** | **YES** | Specifies the name of the Istio Manager CR. |
| **spec.config.numTrustedProxies** | **NO** | Specifies the number of trusted proxies, [more info here](https://istio.io/latest/docs/ops/configuration/traffic-management/network-topologies/). |
