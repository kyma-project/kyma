---
title: Permission-controller chart
type: Configuration
---

To configure the Permission-controller chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

The following table lists the configurable parameters of the permission-controller chart and their default values.

| Parameter | Description | Default value |
| --------- | ----------- | ------------- |
| `global.users.namespaceAdmin.groups` | Comma-separated list of user groups whose members are to be granted admin privileges in all Namespaces except those specified in `config.namespaceBlacklist`. | `"namespace-admins"` |
| `config.namespaceBlacklist` | Comma-separated list of Namespaces that will not be accessible by members of the user groups specified in `global.users.namespaceAdmin.groups`. | `"kyma-system, istio-system, default, knative-eventing, knative-serving, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss"` |
| `config.enableStaticUser`| Boolean that determines whether to grant admin privileges to the `namespace.admin@kyma.cx` static user for testing purposes. | `true` |

- To configure the list of Namespace admin groups, create a Configmap with a global override.
- The remaining parameters can by modified in the chart's `values.yaml` file or by creating a Configmap with component-specific overrides.

>**NOTE:** Members of the groups specified in `global.users.namespaceAdmin.groups` are allowed to create new Namespaces and delete Namespaces they manage.