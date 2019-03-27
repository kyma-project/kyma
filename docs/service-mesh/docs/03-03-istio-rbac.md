---
title: Istio RBAC
type: Details
---

As a core component, Istio is installed with Kyma by default. As a part of this installation, we create a [Cluster RBAC Config](https://istio.io/docs/reference/config/authorization/istio.rbac.v1alpha1/) object, which dictates the global (cluster wide) behavior of Istio. 

The default configuration file can be found [here](https://github.com/kyma-project/kyma/blob/master/resources/core/charts/istio-rbac/templates/rbac-config.yaml)

> **NOTE:** As of Istio 1.1 the previous implementation (RbacConfig) has been deprecated in favor of ClusterRbacConfig. More information can be found in the Istio upgrade [documentation](https://istio.io/docs/setup/kubernetes/upgrade/steps/#migrating-from-rbacconfig-to-clusterrbacconfig)

> **NOTE:** The ClusterRBACConfig object is a singleton, meaning there can be only one configuration file, and only the name `default` is valid. 

## Overriding the default configuration
Because the object is a singleton, any customization of the RBACConig needs to be made in the `default` configuration file. 
Alternatively, it is possible to delete our default configuration, and supply a new one. 

> **NOTE:** As the default configuration is installed in the `kyma-system` namespace, be aware that only users with permissions inside the namespace will have access to edit the resource. By default that would require `kyma-admin` access or `kyma-developer` bound to the `kyma-system` namespace access. 
