---
title: Permission-controller chart
type: Configuration
---

To configure the Permission-controller chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Type | Description | Default value |
| --------- | ---- | ----------- | ------------- |
| `config.subjectGroups` | []string | List of user groups whose members are to be granted admin privileges in all Namespaces, except those specified in `config.namespaceBlacklist`. | ["namespace-admins"] |
| `config.namespaceBlacklist` | []string | List of Namespaces that will not be accessible by members of the user groups specified in `config.subjectGroups`. | ["kyma-system, istio-system, default, knative-eventing, knative-serving, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss"] |
| `config.enableStaticUser`| boolean | Determine whether to grant admin privileges to the `namespace.admin@kyma.cx` static user. | true |

>**NOTE:** To allow members of the subject groups to create Namespaces, make sure each group is bound to the `kyma-namespace-admin` ClusterRole. 