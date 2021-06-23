---
title: Alertmanager sub-chart
---

To configure the Alertmanager sub-chart, override the default values of its `values.yaml` file.
Learn how it works under [Configurable Parameters](./README.md).

Here are some of the parameters you can set. 
For the complete list, see the `values.yaml` file.

## Configurable parameters

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.alertTools.credentials.slack.apiurl** | Specifies the URL endpoint which sends alerts triggered by Prometheus rules.  | None |
| **global.alertTools.credentials.slack.channel** | Refers to the Slack channel which receives notifications on new alerts. | None |
| **global.alertTools.credentials.slack.matchExpression** | Notifications will be sent only for those alerts whose labels match the specified expression.  | "severity: critical" |
| **global.alertTools.credentials.slack.sendResolved** | Specifies whether or not to notify about resolved alerts.  | true |
| **global.alertTools.credentials.victorOps.routingkey** | Defines the team routing key in [VictorOps](https://help.victorops.com/). | None |
| **global.alertTools.credentials.victorOps.apikey** | Defines the team API key in VictorOps. | None |
| **global.alertTools.credentials.victorOps.matchExpression** | Notifications will be sent only for those alerts whose labels match the specified expression.  | "severity: critical" |
| **global.alertTools.credentials.victorOps.sendResolved** | Specifies whether or not to notify about resolved alerts.  | true |

>**NOTE:** Override all configurable values for the Alertmanager sub-chart using Secrets (`kind: Secret`).
