---
title: PodPreset sub-chart
type: Configuration
---

To configure the PodPreset sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **controller.enabled** | Enables the controller-manager which restarts Deployments whenever the PodPreset changes. | `false` |
| **webhook.verbosity** | Defines log severity level. The possible values range from 0-10. | `6` |
