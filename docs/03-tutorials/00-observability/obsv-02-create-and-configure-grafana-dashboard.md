---
title: Create a Grafana dashboard
---

Kyma comes with a set of dashboards for monitoring Kubernetes clusters. These dashboards display metrics that the Prometheus server collects.

You can create and configure a basic Grafana dashboard of a [Gauge](https://grafana.com/docs/grafana/latest/panels/visualizations/gauge-panel/#gauge-panel) type. On the dashboard, you see how the values of the `cpu_temperature_celsius` metric change in time, representing the current processor temperature ranging from 60 to 90 degrees Celsius. The dashboard shows explicitly when the CPU temperature exceeds the pre-defined threshold of 75 degrees Celsius.

In addition to creating a dashboard during runtime, you can also add it to the Kubernetes resources in your repository. This way, the dashboard configuration is portable and you can deploy it together with the application in any new cluster.

## Prerequisites

You have performed the steps to observe application metrics using the `monitoring-custom-metrics` example and successfully deployed the `sample-metrics-8081` service that exposes the `cpu_temperature_celsius` metric.

Follow these sections to create the Gauge dashboard type for the `cpu_temperature_celsius` metric.

## Create the dashboard

1. [Access Grafana](../../04-operation-guides/operations/obsv-02-access-expose-kiali-grafana.md).
2. Add a new dashboard with a new panel.
3. For your new query, select **Prometheus** from the data source selector.
4. Pick the `cpu_temperature_celsius` metric.
5. To retrieve the latest metric value on demand, activate the **Instant** switch.
6. From the visualization panels, select the **Gauge** dashboard type.
7. Save your changes and provide a name for the dashboard.

## Configure the dashboard

1. To edit the dashboard settings, go to the **Panel Title** options and select **Edit**.
2. Find the **Field** tab and set the measuring unit to Celsius degrees, indicating the metric data type.
3. Set the minimum metric value to `60` and the maximum value to `90`, indicating the `cpu_temperature_celsius` metric value range.
4. For the dashboard to turn red once the CPU temperature reaches and exceeds 75°C, set a red color threshold to `75`.
5. Go to the **Panel** tab and title the dashboard, for example, `CPU Temperature`.
6. To display this range on the dashboard, make sure that under **Panel > Display**, the threshold labels and threshold markers are activated.
7. Save your changes. We recommend that you add a note to describe the changes made.

## Check the dashboard

Refresh the browser to see how the dashboard changes according to the current value of the `cpu_temperature_celsius` metric.

- If the current metric value ranges from 60 to 74 degrees Celsius, it turns **green**.
- If the current metric value ranges from 75 to 90 degrees Celsius, it turns **red**.

## Add the dashboard as Kubernetes resource

When you create a dashboard to monitor one of your applications (Function, microservice,...), we recommend that you define the dashboard as a Kubernetes ConfigMap resource. In this case, a Grafana sidecar automatically loads the Dashboard on Grafana startup. Following that approach, you can easily keep the dashboard definition together with the kubernetes resource definitions of your application and port it to different clusters.

1. Create a JSON document with the dashboard definition; for example, by exporting it from Grafana.
2. Create a Kubernetes resource with a unique name for your dashboard and the JSON content, like the following example:

   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: {UNIQUE_DASHBOARD_NAME}-grafana-dashboard
     labels:
       grafana_dashboard: "1"
   data:
     {UNIQUE_DASHBOARD_NAME}-dashboard.json: |-
       {
         # dashboard JSON content
       }
   ```

3. To see your new dashboard, restart the Grafana UI.

## Next steps

- Follow the tutorial to [Define alerting rules](obsv-03-define-alerting-rules-monitor.md).
- If you don't want to proceed with the following tutorial, [clean up the configuration](obsv-05-clean-up-configuration.md).
