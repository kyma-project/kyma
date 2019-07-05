---
title: Cluster Users chart
type: Configuration
---

To configure the Cluster Users chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **bindings.kymaEssentials.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-essentials** ClusterRole. | `[]` |
| **bindings.kymaView.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-view** ClusterRole. | `[]` |
| **bindings.kymaEdit.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-edit** ClusterRole. | `[]` |
| **bindings.kymaAdmin.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-admin** ClusterRole. | `[]` |
| **bindings.kymaDeveloper.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-developer** ClusterRole. | `[]` |
| **users.adminGroup** | Specifies the name of the group used in ClusterRoleBinding to the **kyma-admin** ClusterRole. | `""` |
