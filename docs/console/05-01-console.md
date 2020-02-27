---
title: Console chart
type: Configuration
---

To configure the Console chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

> **TIP:** To learn more about how to use overrides in Kyma, see the following documents:
> * [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
> * [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **cluster.orgId** |  Defines the ID of the organization for which the Kyma cluster is installed. It shows under **General settings** in the Console UI. | `my-org-123` |
| **cluster.orgName** | Defines the name of the organization for which the Kyma cluster is installed. It shows under **General settings** in the Console UI. | `My Organization` |
| **cluster.headerLogoUrl** | Defines the address of the logo image that shows in the Console UI navigation header. | `assets/logo.svg` |
| **cluster.faviconUrl** | Defines the icon that shows in the address bar of a browser. | `favicon.ico` |
| **cluster.headerTitle** | Defines an additional title that shows next to the logo image in the Console UI navigation header. | None |
| **cluster.disabledNavigationNodes** | Defines a list of categories or specific nodes in the navigation that you want to hide in the Console UI navigation. To hide all navigation nodes from a category, make sure the list includes **categoryLabel**. To hide a specific navigation node, use both **categoryLabel** and **nodeLabel** separated with a period (`.`). For the Namespace-related views, add the `namespace` prefix to the list entry. For example, `namespace.operation.secrets` would hide the **Secrets** navigation node within the **Operation** category for all Namespaces. For all labels, use lowercase and don't use any spaces or dashes. | None |
