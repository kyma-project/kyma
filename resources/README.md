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

To monitor the health of your component properly, make sure you include configuration files for ServiceMonitors, alert rules, and dashboards under your component's chart. 

For reference, see [this](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates) example of the service catalog component including the [ServiceMonitor](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates/controller-manager-service-monitor.yaml) and [dashboard](https://github.com/kyma-project/kyma/blob/master/resources/service-catalog/charts/catalog/templates/dashboard-configmap.yaml) configurations.
When adding configuration files, follow this naming convention:

* Use `service-monitor.yaml` for ServiceMonitors.
* Use `dashboard-configmap.yaml`for dashboards.


For details on observing metrics, creating dashboards, and setting alerting rules, see [these](https://kyma-project.io/docs/components/monitoring/#tutorials-tutorials).
