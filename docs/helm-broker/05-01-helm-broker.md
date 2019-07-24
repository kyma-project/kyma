---
title: Helm Broker chart
type: Configuration
---

To configure the Helm Broker chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.cfgReposUrlName** | Specifies the name of the default ConfigMap which provides the URLs of addons repositories. | `helm-repos-urls` |
| **global.isDevelopMode** | Defines that each repository URL must be an HTTPS server. If set to `true`, HTTP servers are also acceptable.  | `false` |
