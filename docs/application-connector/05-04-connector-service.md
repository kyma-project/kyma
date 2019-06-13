---
title: Connector Service sub-chart
type: Configuration
---

To configure the Connector Service sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **deployment.args.tokenLength**| Specifies a number of characters of a registration token. | `64` |
| **deployment.args.appTokenExpirationMinutes** | Specifies a period of time after which a token for an Application expires. It is provided in minutes. | `5` |
| **deployment.args.runtimeTokenExpirationMinutes** | Specifies a period of time after which a token for a runtime expires. It is provided in minutes. | `10` |
| **deployment.args.appValidityTime** | Specifies a period of time during which certificates that the service issues for an Application are valid. It is provided in days. | `90d` |
| **deployment.args.runtimeValidityTime** | Specifies a period of time during which certificates that the service issues for a runtime are valid. It is provided in days. | `90d` |
| **deployment.args.central** | Determines whether the Connector Service works in the central mode. | `&central false` |
| **deployment.envvars.country** | Specifies a country, using two-letter country codes, which is required to generate a Certificate Signing Request. | `DE` |
| **deployment.envvars.organization** | Specifies an organization which is required to generate a Certificate Signing Request. | `Organization` |
| **deployment.envvars.organizationalunit** | Specifies an organization unit which is required to generate a Certificate Signing Request. | `OrgUnit` |
| **deployment.envvars.locality** | Specifies a locality which is required to generate a Certificate Signing Request. | `Waldorf` |
| **deployment.envvars.province** | Specifies a province which is required to generate a Certificate Signing Request. | `Waldorf` |
