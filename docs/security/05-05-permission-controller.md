---
title: Permission Controller chart
type: Configuration
---

To configure the Permission Controller chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

The following table lists the configurable parameters of the permission-controller chart and their default values.

| Parameter | Description | Default value |
| --------- | ----------- | ------------- |
| `global.kymaRuntime.namespaceAdminGroup` | Determines the user group for which a RoleBinding to the **kyma-namespace-admin** role is created in all Namespaces except those specified in the `config.namespaceBlacklist` parameter. | `runtimeNamespaceAdmin` |
| `config.namespaceBlacklist` | Comma-separated list of Namespaces in which a RoleBinding to the **kyma-namespace-admin** role is not created for the members of the group specified in the `global.kymaRuntime.namespaceAdminGroup` parameter.|`kyma-system, istio-system, default, knative-eventing, knative-serving, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss, tekton-pipelines` |
| `config.enableStaticUser`| Determines if a RoleBinding to the **kyma-namespace-admin** role for the static `namespace.admin@kyma.cx` user is created for every Namespace that is not blacklisted. | `true` |
