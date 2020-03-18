---
title: Grafana sub-chart
type: Configuration
---

To configure the Grafana sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can set.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **env.GF_USERS_DEFAULT_THEME** | Specifies the background color of the Grafana UI. You can change it to `dark`. | `light` |
| **env.GF_USERS_AUTO_ASSIGN_ORG_ROLE** | Specifies the auto-assigned [user role](https://grafana.com/docs/grafana/latest/installation/configuration/#users). You can change the value to `Viewer` or `Admin`. | `Editor' |
| **env.GF_LOG_LEVEL** | Specifies the log level used by Grafana. Be aware that logs at the `info` level may print logins, which can potentially be users' email addresses. | `warn` |
| **persistence.enabled** | Specifies whether user and dashboard data used by Grafana is durably persisted. If enabled, the Grafana database will be mounted to a PersistentVolume and survive restarts. | `true` |
