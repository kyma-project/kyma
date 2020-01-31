# Resources                                                                                  

## Overview

Resources are all components in Kyma that are available for local and cluster installation. You can find more details about each component in the corresponding README.md files.

Resources currently include, but are not limited to, the following:

- Elements which are essential for the installation of `core` components in Kyma, such as certificates, users, and permissions
- Examples of the use of specific components
- Scripts for the installation of Helm, Istio deployment, as well as scripts for validating Pods, starting Kyma, and testing

## Development

Every component, test, or tool in the `kyma` repository contains a Makefile. A Makefile is used to build an image of a given component and to push it to the external repository. Every time you create a new component, test, or tool, ensure that its Makefile contains a path to the `values.yaml` file which informs about the actual image version used in Kyma.
To do so, add this entry to the Makefile:

```
.PHONY: path-to-referenced-charts
path-to-referenced-charts:
    @echo "{path to the referenced charts}"
```

The version of the actual component image is located under the **global.{name_of_component}.version** property.
**{name_of_component}** is a directory name of the component where dashes are replaced by underscores.

### Add monitoring to components

To monitor the health of your component properly, make sure you include configuration files for Service Monitors, alert rules, and dashboards under your component's chart. 

For reference, see [this](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates) example of the service catalog component including the [ServiceMonitor](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates/controller-manager-service-monitor.yaml) and [dashboard](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates/dashboard-configmap.yaml) configurations.


When creating a service monitor resource, follow this naming convention:

* To specify the file name, use either `service-monitor.yaml` or `{component_name}-service-monitor.yaml`.
* To specify the resource name in **metadata** section of the file, use the `{chart_name}-{name_of_monitored_chart_component}` pattern . For example, when creating a service monitor resource for Grafana, write `monitoring-grafana`, where the name of the main chart is **monitoring**, and the monitored component is **grafana**. 

When creating alert rule resources, follow this naming convention:

* To specify the file name, use `prometheus-rules.yaml`.
* To specify the resource name in the **metadata** section of the file, use the `{chart_name}` pattern if the resource contains all rules for all the components of the chart, otherwise use `{name_of_main_chart}-{name_of_chart_component}`. For example, write `monitoring` if the resource contains all alert rules for the monitoring component, or `monitoring-grafana`, if it contains just the rules for the grafana component.

When creating dashboard resources, follow this naming convention:

* To specify the file name, use `dashboard-configmap.yaml`.
* To specify the resource name in the **metadata** section of the file, use the `{chart_name}-dashboard` pattern. For example, write `backup-dashboard`. If you create a dashboard for a subchart, use the `{chart_name}-{sub_chart_name}-dashboard` pattern. For example, write `rafter-asyncapi-service-dashboard`. 

For details on observing metrics, creating dashboards, and setting alerting rules, see [these](https://kyma-project.io/docs/components/monitoring/#tutorials-tutorials) tutorials.
