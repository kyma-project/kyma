---
title: Runtime Provisioner chart
type: Configuration
---

To configure the Runtime Provisioner chart, override the default values of its `values.yaml` file. This document describes the parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see [Helm overrides for Kyma installation](root/kyma#configuration-helm-overrides-for-kyma-installation).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **database.schemaFilePath** | Filepath for the database schema | `assets/database/provisioner.sql` |
| **gardener.project** | Name of the Gardener project connected to the service account | `` |
| **gardener.kubeconfig** | Base64-encoded Gardener service account key | `` |
| **provisioner** | Provisioning mechanism used by the Runtime Provisioner (Gardener or Hydroform)  | `gardener` |
| **installation.timeout** | Kyma installation timeout | `30m` |
| **installation.errorsCountFailureThreshold** | Number of installation errors that causes installation to fail | `5` |