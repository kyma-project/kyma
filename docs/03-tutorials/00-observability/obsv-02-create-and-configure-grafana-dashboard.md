---
title: Create a Grafana dashboard
---

This tutorial shows how to create and configure a basic Grafana dashboard of a [Gauge](https://grafana.com/docs/grafana/latest/panels/visualizations/gauge-panel/#gauge-panel) type. The dashboard shows how the values of the `cpu_temperature_celsius` metric change in time, representing the current processor temperature ranging from 60 to 90 degrees Celsius. The dashboard shows explicitly when the CPU temperature exceeds the pre-defined threshold of 75 degrees Celsius.

## Prerequisites

You have performed the steps to observe application metrics using the `monitoring-custom-metrics` example and successfully deployed the `sample-metrics-8081` service which exposes the `cpu_temperature_celsius` metric.

Follow these sections to create the Gauge dashboard type for the `cpu_temperature_celsius` metric.

## Create the dashboard

1. [Access Grafana](../../../04-operation-guides/operations/obsv-02-access-expose-kiali-grafana.md).

2. Add a new dashboard with a new panel.

3. For your new query, select **Prometheus** from the data source selector.

4. Pick the `cpu_temperature_celsius` metric.

5. To retrieve the latest metric value on demand, activate the **Instant** switch.

6. From the visualization panels, select the **Gauge** dashboard type.

7. Save your changes and provide a name for the dashboard.

## Configure the dashboard

1. To edit the dashboard settings, go to the **Panel Title** options and select **Edit**.

2. Find the **Field** tab and set the measuring unit to Celsius degrees to reflect the metric data type.

3. Set the minimum metric value to `60` and the maximum value to `90` to reflect the `cpu_temperature_celsius` metric value range.

4. For the dashboard to turn red once the CPU temperature reaches and exceeds 75 °C, set a red color threshold to `75`.

5. Go to the **Panel** tab and title the dashboard, for example, `CPU Temperature`.

6. Under **Panel > Display**, make sure the threshold labels and threshold markers are activated to display this range on the dashboard.

7. Save your changes. We recommend that you add a note to describe the changes made.

## Verify the dashboard

Refresh the browser to see how the dashboard changes according to the current value of the `cpu_temperature_celsius` metric.

- It turns **green** if the current metric value ranges from 60 to 74 degrees Celsius:

- It turns **red** if the current metric value ranges from 75 to 90 degrees Celsius:

## Next steps

- Follow the tutorial to [Define alerting rules](obsv-04-define-alerting-rules-monitor.md).
- You can also define the dashboard's ConfigMap and add it to the `resources` folder under the given component's chart. To make the dashboard visible, simply use the `kubectl apply` command to deploy it. For details on adding monitoring to components, see the [`README.md`](https://github.com/kyma-project/kyma/blob/master/resources/monitoring/charts/grafana/README.md) document.
- If you don't want to proceed with the following tutorial, [clean up the configuration](obsv-06-clean-up-configuration.md).
