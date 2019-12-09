---
title: Send notifications to Slack
type: Tutorials
---
This tutorial shows you how to configure Alertmanager to send notifications. Alertmanager supports several [notification receivers](https://prometheus.io/docs/alerting/configuration/#receiver), but this tutorial only focuses on sending notifications to Slack.

## Prerequisites

This  tutorial is a follow-up of the [**Observe application metrics**](#tutorials-observe-application-metrics) and the [**Define alerting rules**](#tutorials-define-alerting-rules) tutorials that use the `monitoring-custom-metrics` example. Follow this tutorial to deploy the `sample-metrics-8081` service which exposes the `cpu_temperature_celsius` metric and creates an alert based on it. That configuration is required to complete this tutorial.

## Steps

Follow these steps to configure notifications for Slack every time Alertmanager triggers and resolves the `CPUTempHigh` alert.

1. Install the Incoming WebHooks application using Slack App Directory.

   >**NOTE**: The approval of your Slack workspace administrator may be necessary to install the application.

2. Configure the application to receive notifications coming from third-party services. Read [this](https://api.slack.com/incoming-webhooks#create_a_webhook) document to find out how to set up the configuration for Slack.

   The integration settings should look similar to the following:

   ![Integration Settings](./assets/integration-settings.png)

3. Override Alertmanager configuration. The configuration for notification receivers is located in [this](https://github.com/kyma-project/kyma/blob/master/resources/monitoring/templates/alertmanager/alertmanager.config.yaml) template. By default, it contains settings for VictorOps, Slack, and Webhooks. Define a Secret to [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation) default [values](https://github.com/kyma-project/kyma/blob/master/resources/monitoring/charts/prometheus-node-exporter/values.yaml) used by the chart.

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
       global.alertTools.credentials.slack.apiurl: "{WEBHOOK_URL}"
   ```

   Use the following parameters:

   | Parameter | Description |
   |-----------|--------------------|
   | **global.alertTools.credentials.slack.channel** | Specifies the Slack channel which receives notifications on new alerts, such as `test-monitoring-alerts`.
   | **global.alertTools.credentials.slack.apiurl** | Specifies the URL endpoint which sends alerts triggered by Prometheus rules. The Incoming WebHooks application provides you with the Webhook URL, such as `https://hooks.slack.com/services/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`, that you can paste in this configuration. |

   For details on Alertmanager chart configuration and parameters, see [this](https://kyma-project.io/docs/master/components/monitoring/#details-alertmanager-alertmanager-configuration) document.

4. Deploy the Secret. Use this command:

   ```bash
   kubectl apply -f {FILE_NAME}.yaml
   ```

5. Proceed with Kyma installation.

   >**NOTE**: If you add the overrides in the runtime, trigger the update process using this command:
   >```
   >kubectl label installation/kyma-installation action=install
   >```

6. Verify if your Slack channel receives alert notifications about firing and resolved alerts. See the example:

   ![Alert Notifications](./assets/alert-notifications.png)
