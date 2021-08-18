---
title: Cluster Users chart
---

To configure the Cluster Users chart, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/cluster-users/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/04-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **bindings.kymaEssentials.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-essentials** ClusterRole. | `[]` |
| **bindings.kymaView.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-view** ClusterRole. | `[]` |
| **bindings.kymaEdit.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-edit** ClusterRole. | `[]` |
| **bindings.kymaAdmin.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-admin** ClusterRole. | `[]` |
| **bindings.kymaDeveloper.groups** | Specifies the array of groups used in ClusterRoleBinding to the **kyma-developer** ClusterRole. | `[]` |
| **users.administrators** | Specifies the array of names used in ClusterRoleBinding to the **kyma-admin** ClusterRole. | `["admin@kyma.cx"]` |
| **users.adminGroup** | Specifies the name of the group used in ClusterRoleBinding to the **kyma-admin** ClusterRole. | `""` |
