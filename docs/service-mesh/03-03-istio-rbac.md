---
title: Istio RBAC configuration
type: Details
---

As a core component, Istio is installed with Kyma by default. The [ClusterRbacConfig](https://istio.io/docs/reference/config/security/istio.rbac.v1alpha1/) custom resource (CR), which defines the global behavior of Istio, is created as a part of the installation process.

The default Istio RBAC configuration is defined in [this](https://github.com/kyma-project/kyma/blob/master/resources/core/charts/istio-rbac/templates/rbac-config.yaml) file. 

## Override the default configuration

To override the default configuration of Istio RBAC, edit the ClusterRbacConfig CR on a running cluster. This CR is created in the `kyma-system` Namespace and therefore requires admin permissions to edit it. 

To show the current Istio RBAC configuration in the `yaml` format, run: 
```bash
kubectl get -n kyma-system clusterrbacconfig -o yaml
```

To edit the Istio RBAC configuration, run: 
```bash
kubectl edit -n kyma-system clusterrbacconfig
```

> **NOTE:** The `ClusterRbacConfig` object is a singleton, which means that only a single object of this kind can exist in a cluster. Additionally, the only valid name for the object is `default`. As such, the best way to customize Istio RBAC is by editing the existing `ClusterRbacConfig` object. 
