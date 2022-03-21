---
title: Cluster Users chart
---

To configure the Cluster Users chart, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/cluster-users/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **bindings.kymaEssentials.groups** | Specifies the array of groups used in Cluster Role Binding to the **kyma-essentials** Cluster Role. | `[]` |
| **bindings.kymaView.groups** | Specifies the array of groups used in Cluster Role Binding to the **kyma-view** Cluster Role. | `[]` |
| **bindings.kymaEdit.groups** | Specifies the array of groups used in Cluster Role Binding to the **kyma-edit** Cluster Role. | `[]` |
| **bindings.kymaAdmin.groups** | Specifies the array of groups used in Cluster Role Binding to the **kyma-admin** Cluster Role. | `[]` |
| **bindings.kymaDeveloper.groups** | Specifies the array of groups used in Cluster Role Binding to the **kyma-developer** Cluster Role. | `[]` |
| **users.administrators** | Specifies the array of names used in Cluster Role Binding to the **kyma-admin** Cluster Role. | `["admin@kyma.cx"]` |
| **users.adminGroup** | Specifies the name of the group used in Cluster Role Binding to the **kyma-admin** Cluster Role. | `""` |
