---
title: Service Catalog: PodPreset sub-chart
---

To configure the PodPreset sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **controller.enabled** | Enables the controller-manager which restarts Deployments whenever the PodPreset changes. | `false` |
| **webhook.verbosity** | Defines log severity level. The possible values range from 0-10. | `6` |
