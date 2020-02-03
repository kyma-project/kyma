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

Include configuration files for Service Monitors, alert rules, and dashboards under your component's chart to ensure proper health check monitoring of your component. 

For an example of such a component, see [Service Catalog](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates) that contains the [ServiceMonitor](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates/controller-manager-service-monitor.yaml) and [dashboard](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates/dashboard-configmap.yaml) configurations.


When creating a ServiceMonitor resource, follow this naming convention:

| Resource | Description | Name/Pattern | Example |
|-----------|-------------|---------------| --------|
| Service monitor| Name of the file which contains Service Monitor specification. | `service-monitor.yaml`  |`service-monitor.yaml`|
| Service monitor| Name of the resource in the **metadata** section of the file.  | `{chart_name}-{name_of_monitored_chart_component}` | `monitoring-grafana`, where the name of the main chart is **monitoring**, and the monitored component is **grafana**.|
| Alert rule| Name of the file which contains Alert rule specification | `prometheus-rules.yaml` | `prometheus-rules.yaml` |
| Alert rule| Name of the resource in the **metadata** section of the file. | `{chart_name}` if the resource contains rules for all components of the chart, `{name_of_main_chart}-{name_of_chart_component}` if every component has its own set of rules.  |`monitoring` if the resource contains all alert rules for the monitoring component, `monitoring-grafana`, if it contains only the rules for the grafana component. |
| Dashboard| Name of the file which contains dashboard specification. |`dashboard-configmap.yaml`|`dashboard-configmap.yaml`|
| Dashboard| Name of the resource in the **metadata** section of the file.| `{chart_name}-dashboard` for the main chart dashboard,`{chart_name}-{sub_chart_name}-dashboard` for the sub-chart. | `backup-dashboard`, `rafter-asyncapi-service-dashboard`. |  

For details on observing metrics, creating dashboards, and setting alerting rules, see [these](https://kyma-project.io/docs/components/monitoring/#tutorials-tutorials) tutorials.
