---
title: Certificate Manager chart
type: Configuration
---

The Certificate Manager chart contains embedded [cert-manager.io](https://cert-manager.io/) instance.
It should be installed in a dedicated namespace (cert-manager)
To configure the Certificate Manager chart, override the default values of its `values.yaml` file.
This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **modules.manager.enabled** | Controls if cert-manager is enabled. | `true` |

>**NOTE:** For Cert-Manager configuration options suitable for your use-case please refer to cert-manager [documentation](https://cert-manager.io/docs/).
>You can find the original cert-manager configuration file in the following location: `./charts/cert-manager/values.yaml`
>When overriding these values please put the overrides in the root chart `./values.yaml` file.

