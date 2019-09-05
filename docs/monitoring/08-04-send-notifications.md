---
title: Send notifications to Slack
type: Tutorials
---
This tutorial shows you how to configure Alertmanager to send notifications. Alertmanager supports several [notification receivers](https://prometheus.io/docs/alerting/configuration/), however, this tutorial focuses on sending notifications to Slack.

## Prerequisites

This  tutorial is a follow-up of the [observe application metrics](/components/monitoring/#tutorials-observe-application-metrics) and the  [define alerting rules](https://kyma-project.io/docs/master/components/monitoring/#tutorials-define-alerting-rules) tutorials that use the `monitoring-custom-metrics` example. Follow this tutorial to deploy the `sample-metrics-8081` service which exposes the `cpu_temperature_celsius` metric, and creates an alert based on it. That configuration is required to complete this tutorial.


## Steps

Follow these steps to configure notifications for Slack.


1. Install the Incoming Webhooks Slack app and configure it to receive notifications coming from third party services. Read [this](https://api.slack.com/incoming-webhooks#create_a_webhook) document to find out how to set up the configuration. 
  >**NOTE**: The approval of your Slack workspace administrator may be necessary to set up the webhook.

 The integration settings should look similar to the following:

 ![Integration Settings](./assets/integration-settings.png)

2. The configuration for notification receivers is located in [this](https://github.com/kyma-project/kyma/blob/master/resources/monitoring/charts/alertmanager/templates/alertmanager.config.yaml) template. By default, it contains settings for VictorOps, Slack, and Webhooks. Define a Secret to [override](../../../../docs/kyma/05-03-overrides.md) default [values](https://github.com/kyma-project/kyma/blob/master/resources/monitoring/charts/alertmanager/values.yaml) used by the chart.

```yaml
apiVersion: v1
kind: Secret
metadata:
 name: monitoring-config-overrides
 namespace: kyma-installer
 labels:
    kyma-project.io/installation: ""
    installer: overrides
    component: monitoring
type: Opaque
stringData:
    global.alertTools.credentials.slack.channel: "{CHANNEL_NAME}"
    global.alertTools.credentials.slack.apiurl: "{HOOK_ENDPOINT}"
```
Use the following parameters:

| Parameter | Description |
|-----------|--------------------|
| **global.alertTools.credentials.slack.channel** | Specifies the Slack channel which receives notifications on new alerts. It must be the same channel you specified in the Slack webhook configuration. 
| **global.alertTools.credentials.slack.apiurl** | Specifies the URL endpoint which sends alerts triggered by Prometheus rules. The Incoming Webhooks Slack app provides you with the URL when creating the integration.|

For details on Alertmanager chart configuration and parameters see [this](components/monitoring/#configuration-alertmanager-sub-chart) document.

3. Proceed with Kyma installation. 

  >**NOTE**: If you add the overrides in the runtime, trigger the update process using this command:
  >```
  >kubectl label installation/kyma-installation action=install
  >```

4. Verify if your Slack channel receives alert notifications about firing and resolved alerts. See the example:

![Alert Notifications](./assets/alert-notifications.png)



