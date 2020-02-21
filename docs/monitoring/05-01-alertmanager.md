---
title: Alertmanager sub-chart
type: Configuration
---

To configure the Alertmanager sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can set.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.alertTools.credentials.slack.apiurl** | Specifies the URL endpoint which sends alerts triggered by Prometheus rules.  | None |
| **global.alertTools.credentials.slack.channel** | Refers to the Slack channel which receives notifications on new alerts. | None |
| **global.alertTools.credentials.victorOps.routingkey** | Defines the team routing key in [VictorOps](https://help.victorops.com/). | None |
| **global.alertTools.credentials.victorOps.apikey** | Defines the team API key in VictorOps. | None |
| **global.alertTools.credentials.routes.enableWarningAlertToVictorOps** | Defines the team route to VictorOps when warning alert appeared.  | None |
| **global.alertTools.credentials.routes.enableWarningAlertToSlack** | Defines the team route to Slack channel when warning alert appeared.  | None |

>**NOTE:** Override all configurable values for the Alertmanager sub-chart using Secrets (`kind: Secret`).
