---
title: Runtime Provisioner chart
type: Configuration
---

To configure the Runtime Provisioner chart, override the default values of its `values.yaml` file. This document describes the parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see [Helm overrides for Kyma installation](https://github.com/kyma-project/kyma/blob/master/docs/kyma/05-03-overrides.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **SchemaFilePath** | Filepath for the database schema | `assets/database/provisioner.sql` |