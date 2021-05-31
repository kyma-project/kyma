---
title: Application Connector chart
type: Configuration
---

To configure the Application Connector (AC) chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **central_connectivity_validator.enabled** | Specifies whether to use Central Connectivity Validator for Application Operator. If enabled, it removes the existing per-Application Gateways and installs the Central Connectivity Validator chart. | `false` |