---
title: Istio RBAC
type: Details
---

As a core component, Istio is installed with Kyma by default. As a part of this installation, we create a [Global RBAC Config](https://istio.io/docs/reference/config/authorization/istio.rbac.v1alpha1/) object, which dictates the global (cluster wide) behavior of Istio. 

The default configuration file can be found in the `kyma-system` namespace:

```yaml
---
apiVersion: "rbac.istio.io/v1alpha1"
kind: RbacConfig
metadata:
  name: default
spec:
  mode: 'OFF'
```

> **NOTE:** The RBACConfig object is a singleton, meaning there can be only one configuration file, and only the name `default` is valid. 

## Overriding the default configuration
Because the object is a singleton, any customization of the RBACConig needs to be made in the `default` configuration file. 
Alternatively, it is possible to delete our default configuration, and supply a new one. 
