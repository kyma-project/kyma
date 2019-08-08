---
title: Create a Grafana dashboard
type: Tutorials
---

This tutorial shows how to create and configure a basic Grafana dashboard of a [Gauge](https://grafana.com/docs/features/panels/singlestat/#gauge) type. The dashboard shows how the values of the `cpu_temperature_celsius` metric change in time, representing the current processor temperature ranging from `60` to `90` Celsius degrees. The dashboard shows explicitly when the CPU temperature exceeds the pre-defined threshold of `75` Celsius degrees.

## Prerequisites

This tutorial is a follow-up of the [**Observe application metrics**](#tutorials-observe-application-metrics) tutorial that uses the `monitoring-custom-metrics` example. This example deploys the `sample-metrics-8081` service which exposes the `cpu_temperature_celsius` metric. The configuration is required to complete this tutorial.

## Steps

Access the Grafana UI, select the `cpu_temperature_celsius` metric, the Gauge dashboard type, and start the configuration.

### Create the dashboard

1. Navigate to Grafana. It is available under the `https://grafana.{DOMAIN}` address, where `{DOMAIN}` is the domain of your Kyma cluster, such as `https://grafana.34.63.57.190.xip.io` or `https://grafana.nightly.a.build.kyma-project.io/`. You can also access it by clicking on **Stats & Metrics** on the left navigation menu in the Console UI.

![Stats and Metrics](./assets/stats-and-metrics.png)

2. Click the **+** icon on the left sidebar and select **Dashboard** from the **Create** menu.

![Create a dashboard](./assets/create-dashboard.png)

3. Select **Add Query**.

![Add Query](./assets/add-query.png)

4. Select Prometheus data source from the **Queries to** drop-down list and choose the `cpu_temperature_celsius` metric.

![New dashboard](./assets/new-dashboard.png)

5. Select the **Instant** query to be able to retrieve the last value for each time series.

![Instant option](./assets/instant.png)

6. Switch to the **Visualization** section and select the **Gauge** dashboard type.

![Gauge dashboard type](./assets/gauge-dashboard-type.png)

7. Save the changes by clicking the disk icon in the top right corner of the page. Provide a name for the dashboard.

![Save the dashboard](./assets/save-dashboard.png)

### Configure the dashboard

To edit the dashboard settings, click the **Panel Title** and select **Edit**.

![Edit the dashboard](./assets/edit-dashboard.png)

1. Back in the **Visualization** section, set up the measuring unit to Celsius degrees to reflect the metric data type.

![Temperature](./assets/temperature-celsius.png)

2. Set the minimum metric value to `60` and the maximum value to `90` to reflect the `cpu_temperature_celsius` metric value range. Enable the **Show labels** option to display this range on the dashboard.

![Minimum and maximum values](./assets/min-max-values.png)

3. Set a red color threshold to `75` for the dashboard to turn red once the CPU temperature reaches and exceeds this value.

![Threshold](./assets/threshold.png)

4. Go to the **General** section and give a title to the dashboard.

![Panel title](./assets/panel-title.png)

5. Save the changes by clicking the disk icon in the top right corner of the page. Add an optional note to describe the changes made.

![Note](./assets/save-note.png)

### Test the dashboard

Refresh the settings to see how the dashboard changes according to the current value of the `cpu_temperature_celsius` metric.

- It turns green if the current metric value ranges from `60` to `74` Celsius degrees:

![Green dashboard](./assets/green-dashboard.png)

- It turns green if the current metric value ranges from `75` to `90` Celsius degrees:

![Red dashboard](./assets/red-dashboard.png)
