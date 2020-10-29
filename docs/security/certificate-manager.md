---
title: Certificate Manager chart
type: Configuration
---

The Certificate Manager chart contains embedded [cert-manager.io](https://cert-manager.io/) instance.
It should work in a dedicated namespace (cert-manager)
To configure the Certificate Manager chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **modules.manager.enabled** | Controls if cert-manager is enabled. | `true` |
