---
title: Asset Store Metadata Service sub-chart
type: Configuration
---

To configure the Asset Store Metadata Service sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **replicaCount** | Defines the number of replicas of the service. | `1` |
| **virtualservice.enabled** |       | `false` |
| **virtualservice.annotations** |       | `{}` |
| **virtualservice.name** |       | `asset-metadata-service` |
